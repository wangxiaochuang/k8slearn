package dynamiccertificates

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type DynamicCertKeyPairContent struct {
	name        string
	certFile    string
	keyFile     string
	certKeyPair atomic.Value
	listeners   []Listener
	queue       workqueue.RateLimitingInterface
}

var _ CertKeyContentProvider = &DynamicCertKeyPairContent{}
var _ ControllerRunner = &DynamicCertKeyPairContent{}

func NewDynamicServingContentFromFiles(purpose, certFile, keyFile string) (*DynamicCertKeyPairContent, error) {
	if len(certFile) == 0 || len(keyFile) == 0 {
		return nil, fmt.Errorf("missing filename for serving cert")
	}
	name := fmt.Sprintf("%s::%s::%s", purpose, certFile, keyFile)

	ret := &DynamicCertKeyPairContent{
		name:     name,
		certFile: certFile,
		keyFile:  keyFile,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), fmt.Sprintf("DynamicCABundle-%s", purpose)),
	}
	if err := ret.loadCertKeyPair(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *DynamicCertKeyPairContent) AddListener(listener Listener) {
	c.listeners = append(c.listeners, listener)
}

func (c *DynamicCertKeyPairContent) loadCertKeyPair() error {
	cert, err := ioutil.ReadFile(c.certFile)
	if err != nil {
		return err
	}
	key, err := ioutil.ReadFile(c.keyFile)
	if err != nil {
		return err
	}
	if len(cert) == 0 || len(key) == 0 {
		return fmt.Errorf("missing content for serving cert %q", c.Name())
	}

	// Ensure that the key matches the cert and both are valid
	_, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return err
	}

	newCertKey := &certKeyContent{
		cert: cert,
		key:  key,
	}

	// check to see if we have a change. If the values are the same, do nothing.
	existing, ok := c.certKeyPair.Load().(*certKeyContent)
	if ok && existing != nil && existing.Equal(newCertKey) {
		return nil
	}

	c.certKeyPair.Store(newCertKey)
	klog.V(2).InfoS("Loaded a new cert/key pair", "name", c.Name())

	for _, listener := range c.listeners {
		listener.Enqueue()
	}

	return nil
}

func (c *DynamicCertKeyPairContent) RunOnce(ctx context.Context) error {
	return c.loadCertKeyPair()
}

func (c *DynamicCertKeyPairContent) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.InfoS("Starting controller", "name", c.name)
	defer klog.InfoS("Shutting down controller", "name", c.name)

	// 一直执行，没秒检测一次是否结束
	go wait.Until(c.runWorker, time.Second, ctx.Done())

	// 一直执行，每分钟检测一次是否结束
	go wait.Until(func() {
		if err := c.watchCertKeyFile(ctx.Done()); err != nil {
			klog.ErrorS(err, "Failed to watch cert and key file, will retry later")
		}
	}, time.Minute, ctx.Done())
}

func (c *DynamicCertKeyPairContent) watchCertKeyFile(stopCh <-chan struct{}) error {
	// 预先加一个
	c.queue.Add(workItemKey)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating fsnotify watcher: %v", err)
	}
	defer w.Close()

	if err := w.Add(c.certFile); err != nil {
		return fmt.Errorf("error adding watch for file %s: %v", c.certFile, err)
	}
	if err := w.Add(c.keyFile); err != nil {
		return fmt.Errorf("error adding watch for file %s: %v", c.keyFile, err)
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

func (c *DynamicCertKeyPairContent) handleWatchEvent(e fsnotify.Event, w *fsnotify.Watcher) error {
	defer c.queue.Add(workItemKey)
	// 只处理删除和改名
	if e.Op&(fsnotify.Remove|fsnotify.Rename) == 0 {
		return nil
	}
	// 先删除，防止重复添加
	if err := w.Remove(e.Name); err != nil {
		klog.InfoS("Failed to remove file watch, it may have been deleted", "file", e.Name, "err", err)
	}
	if err := w.Add(e.Name); err != nil {
		return fmt.Errorf("error adding watch for file %s: %v", e.Name, err)
	}
	return nil
}

func (c *DynamicCertKeyPairContent) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *DynamicCertKeyPairContent) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	err := c.loadCertKeyPair()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

func (c *DynamicCertKeyPairContent) Name() string {
	return c.name
}

func (c *DynamicCertKeyPairContent) CurrentCertKeyContent() ([]byte, []byte) {
	certKeyContent := c.certKeyPair.Load().(*certKeyContent)
	return certKeyContent.cert, certKeyContent.key
}
