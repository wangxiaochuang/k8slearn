package transport

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"reflect"
	"sync"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/util/connrotation"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const workItemKey = "key"

var CertCallbackRefreshDuration = 5 * time.Minute

type reloadFunc func(*tls.CertificateRequestInfo) (*tls.Certificate, error)

type dynamicClientCert struct {
	clientCert *tls.Certificate
	certMtx    sync.RWMutex

	reload     reloadFunc
	connDialer *connrotation.Dialer

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface
}

func certRotatingDialer(reload reloadFunc, dial utilnet.DialFunc) *dynamicClientCert {
	d := &dynamicClientCert{
		reload:     reload,
		connDialer: connrotation.NewDialer(connrotation.DialFunc(dial)),
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DynamicClientCertificate"),
	}

	return d
}

func (c *dynamicClientCert) loadClientCert() (*tls.Certificate, error) {
	cert, err := c.reload(nil)
	if err != nil {
		return nil, err
	}

	// 加读锁，如果相同就直接返回
	c.certMtx.RLock()
	haveCert := c.clientCert != nil
	if certsEqual(c.clientCert, cert) {
		c.certMtx.RUnlock()
		return c.clientCert, nil
	}
	c.certMtx.RUnlock()

	// 不同，需要设置新的
	c.certMtx.Lock()
	c.clientCert = cert
	c.certMtx.Unlock()

	if !haveCert {
		return cert, nil
	}

	// 证书变了，要关闭老的连接
	klog.V(1).Infof("certificate rotation detected, shutting down client connections to start using new credentials")
	c.connDialer.CloseAll()

	return cert, nil
}

func certsEqual(left, right *tls.Certificate) bool {
	if left == nil || right == nil {
		return left == right
	}

	if !byteMatrixEqual(left.Certificate, right.Certificate) {
		return false
	}

	if !reflect.DeepEqual(left.PrivateKey, right.PrivateKey) {
		return false
	}

	if !byteMatrixEqual(left.SignedCertificateTimestamps, right.SignedCertificateTimestamps) {
		return false
	}

	if !bytes.Equal(left.OCSPStaple, right.OCSPStaple) {
		return false
	}

	return true
}

func byteMatrixEqual(left, right [][]byte) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if !bytes.Equal(left[i], right[i]) {
			return false
		}
	}
	return true
}

func (c *dynamicClientCert) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(3).Infof("Starting client certificate rotation controller")
	defer klog.V(3).Infof("Shutting down client certificate rotation controller")

	go wait.Until(c.runWorker, time.Second, stopCh)

	go wait.PollImmediateUntil(CertCallbackRefreshDuration, func() (bool, error) {
		c.queue.Add(workItemKey)
		return false, nil
	}, stopCh)

	<-stopCh
}

func (c *dynamicClientCert) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *dynamicClientCert) processNextWorkItem() bool {
	// 获取一个开始处理
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	_, err := c.loadClientCert()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

func (c *dynamicClientCert) GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return c.loadClientCert()
}
