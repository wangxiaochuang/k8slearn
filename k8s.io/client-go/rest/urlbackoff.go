package rest

import (
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"
)

var serverIsOverloadedSet = sets.NewInt(429)
var maxResponseCode = 499

type BackoffManager interface {
	UpdateBackoff(actualUrl *url.URL, err error, responseCode int)
	CalculateBackoff(actualUrl *url.URL) time.Duration
	Sleep(d time.Duration)
}

type URLBackoff struct {
	Backoff *flowcontrol.Backoff
}

type NoBackoff struct {
}

func (n *NoBackoff) UpdateBackoff(actualUrl *url.URL, err error, responseCode int) {
	// do nothing.
}

func (n *NoBackoff) CalculateBackoff(actualUrl *url.URL) time.Duration {
	return 0 * time.Second
}

func (n *NoBackoff) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (b *URLBackoff) Disable() {
	klog.V(4).Infof("Disabling backoff strategy")
	b.Backoff = flowcontrol.NewBackOff(0*time.Second, 0*time.Second)
}

func (b *URLBackoff) baseUrlKey(rawurl *url.URL) string {
	host, err := url.Parse(rawurl.String())
	if err != nil {
		klog.V(4).Infof("Error extracting url: %v", rawurl)
		panic("bad url!")
	}
	return host.Host
}

func (b *URLBackoff) UpdateBackoff(actualUrl *url.URL, err error, responseCode int) {
	// range for retry counts that we store is [0,13]
	if responseCode > maxResponseCode || serverIsOverloadedSet.Has(responseCode) {
		b.Backoff.Next(b.baseUrlKey(actualUrl), b.Backoff.Clock.Now())
		return
	} else if responseCode >= 300 || err != nil {
		klog.V(4).Infof("Client is returning errors: code %v, error %v", responseCode, err)
	}

	//If we got this far, there is no backoff required for this URL anymore.
	b.Backoff.Reset(b.baseUrlKey(actualUrl))
}

func (b *URLBackoff) CalculateBackoff(actualUrl *url.URL) time.Duration {
	return b.Backoff.Get(b.baseUrlKey(actualUrl))
}

func (b *URLBackoff) Sleep(d time.Duration) {
	b.Backoff.Clock.Sleep(d)
}
