package dynamiccertificates

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

var FileRefreshDuration = 1 * time.Minute

type ControllerRunner interface {
	RunOnce(ctx context.Context) error
	Run(ctx context.Context, workers int)
}

type DynamicFileCAContent struct {
	name      string
	filename  string
	caBundle  atomic.Value
	listeners []Listener
	queue     workqueue.RateLimitingInterface
}

var _ Notifier = &DynamicFileCAContent{}
var _ CAContentProvider = &DynamicFileCAContent{}
var _ ControllerRunner = &DynamicFileCAContent{}

type caBundleAndVerifier struct {
	caBundle      []byte
	verifyOptions x509.VerifyOptions
}

func NewDynamicCAContentFromFile(purpose, filename string) (*DynamicFileCAContent, error) {
	if len(filename) == 0 {
		return nil, fmt.Errorf("missing filename for ca bundle")
	}
	name := fmt.Sprintf("%s::%s", purpose, filename)

	ret := &DynamicFileCAContent{
		name:     name,
		filename: filename,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), fmt.Sprintf("DynamicCABundle-%s", purpose)),
	}
	if err := ret.loadCABundle(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *DynamicFileCAContent) AddListener(listener Listener) {
	c.listeners = append(c.listeners, listener)
}

func (c *DynamicFileCAContent) loadCABundle() error {
	caBundle, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return err
	}
	if len(caBundle) == 0 {
		return fmt.Errorf("missing content for CA bundle %q", c.Name())
	}

	// check to see if we have a change. If the values are the same, do nothing.
	if !c.hasCAChanged(caBundle) {
		return nil
	}

	caBundleAndVerifier, err := newCABundleAndVerifier(c.Name(), caBundle)
	if err != nil {
		return err
	}
	c.caBundle.Store(caBundleAndVerifier)
	klog.V(2).InfoS("Loaded a new CA Bundle and Verifier", "name", c.Name())

	for _, listener := range c.listeners {
		listener.Enqueue()
	}

	return nil
}

func (c *DynamicFileCAContent) hasCAChanged(caBundle []byte) bool {
	uncastExisting := c.caBundle.Load()
	if uncastExisting == nil {
		return true
	}

	// check to see if we have a change. If the values are the same, do nothing.
	existing, ok := uncastExisting.(*caBundleAndVerifier)
	if !ok {
		return true
	}
	if !bytes.Equal(existing.caBundle, caBundle) {
		return true
	}

	return false
}

func (c *DynamicFileCAContent) RunOnce(ctx context.Context) error {
	return c.loadCABundle()
}

func (c *DynamicFileCAContent) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.InfoS("Starting controller", "name", c.name)
	defer klog.InfoS("Shutting down controller", "name", c.name)

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, ctx.Done())

	// start the loop that watches the CA file until stopCh is closed.
	go wait.Until(func() {
		if err := c.watchCAFile(ctx.Done()); err != nil {
			klog.ErrorS(err, "Failed to watch CA file, will retry later")
		}
	}, time.Minute, ctx.Done())

	<-ctx.Done()
}

func (c *DynamicFileCAContent) watchCAFile(stopCh <-chan struct{}) error {
	c.queue.Add(workItemKey)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating fsnotify watcher: %v", err)
	}
	defer w.Close()

	if err = w.Add(c.filename); err != nil {
		return fmt.Errorf("error adding watch for file %s: %v", c.filename, err)
	}
	c.queue.Add(workItemKey)

	for {
		select {
		case e := <-w.Events:
			if err := c.handleWatchEvent(e, w); err != nil {
				return err
			}
		case err := <-w.Errors:
			return fmt.Errorf("received fsnotify error: %v", err)
		case <-stopCh:
			return nil
		}
	}
}

func (c *DynamicFileCAContent) handleWatchEvent(e fsnotify.Event, w *fsnotify.Watcher) error {
	// This should be executed after restarting the watch (if applicable) to ensure no file event will be missing.
	defer c.queue.Add(workItemKey)
	if e.Op&(fsnotify.Remove|fsnotify.Rename) == 0 {
		return nil
	}
	if err := w.Remove(c.filename); err != nil {
		klog.InfoS("Failed to remove file watch, it may have been deleted", "file", c.filename, "err", err)
	}
	if err := w.Add(c.filename); err != nil {
		return fmt.Errorf("error adding watch for file %s: %v", c.filename, err)
	}
	return nil
}

func (c *DynamicFileCAContent) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *DynamicFileCAContent) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	err := c.loadCABundle()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

func (c *DynamicFileCAContent) Name() string {
	return c.name
}

func (c *DynamicFileCAContent) CurrentCABundleContent() (cabundle []byte) {
	return c.caBundle.Load().(*caBundleAndVerifier).caBundle
}

func (c *DynamicFileCAContent) VerifyOptions() (x509.VerifyOptions, bool) {
	uncastObj := c.caBundle.Load()
	if uncastObj == nil {
		return x509.VerifyOptions{}, false
	}

	return uncastObj.(*caBundleAndVerifier).verifyOptions, true
}

func newCABundleAndVerifier(name string, caBundle []byte) (*caBundleAndVerifier, error) {
	if len(caBundle) == 0 {
		return nil, fmt.Errorf("missing content for CA bundle %q", name)
	}

	// Wrap with an x509 verifier
	var err error
	verifyOptions := defaultVerifyOptions()
	verifyOptions.Roots, err = cert.NewPoolFromBytes(caBundle)
	if err != nil {
		return nil, fmt.Errorf("error loading CA bundle for %q: %v", name, err)
	}

	return &caBundleAndVerifier{
		caBundle:      caBundle,
		verifyOptions: verifyOptions,
	}, nil
}

func defaultVerifyOptions() x509.VerifyOptions {
	return x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}
