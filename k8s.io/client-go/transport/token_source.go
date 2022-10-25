package transport

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog/v2"
)

func TokenSourceWrapTransport(ts oauth2.TokenSource) func(http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &tokenSourceTransport{
			base: rt,
			ort: &oauth2.Transport{
				Source: ts,
				Base:   rt,
			},
		}
	}
}

type ResettableTokenSource interface {
	oauth2.TokenSource
	ResetTokenOlderThan(time.Time)
}

func ResettableTokenSourceWrapTransport(ts ResettableTokenSource) func(http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &tokenSourceTransport{
			base: rt,
			ort: &oauth2.Transport{
				Source: ts,
				Base:   rt,
			},
			src: ts,
		}
	}
}

func NewCachedFileTokenSource(path string) *cachingTokenSource {
	return &cachingTokenSource{
		now:    time.Now,
		leeway: 10 * time.Second,
		base: &fileTokenSource{
			path:   path,
			period: time.Minute,
		},
	}
}

func NewCachedTokenSource(ts oauth2.TokenSource) *cachingTokenSource {
	return &cachingTokenSource{
		now:  time.Now,
		base: ts,
	}
}

type tokenSourceTransport struct {
	base http.RoundTripper
	ort  http.RoundTripper
	src  ResettableTokenSource
}

var _ utilnet.RoundTripperWrapper = &tokenSourceTransport{}

func (tst *tokenSourceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// This is to allow --token to override other bearer token providers.
	if req.Header.Get("Authorization") != "" {
		return tst.base.RoundTrip(req)
	}
	// record time before RoundTrip to make sure newly acquired Unauthorized
	// token would not be reset. Another request from user is required to reset
	// and proceed.
	start := time.Now()
	resp, err := tst.ort.RoundTrip(req)
	if err == nil && resp != nil && resp.StatusCode == 401 && tst.src != nil {
		tst.src.ResetTokenOlderThan(start)
	}
	return resp, err
}

func (tst *tokenSourceTransport) CancelRequest(req *http.Request) {
	if req.Header.Get("Authorization") != "" {
		tryCancelRequest(tst.base, req)
		return
	}
	tryCancelRequest(tst.ort, req)
}

func (tst *tokenSourceTransport) WrappedRoundTripper() http.RoundTripper { return tst.base }

type fileTokenSource struct {
	path   string
	period time.Duration
}

var _ = oauth2.TokenSource(&fileTokenSource{})

func (ts *fileTokenSource) Token() (*oauth2.Token, error) {
	tokb, err := ioutil.ReadFile(ts.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file %q: %v", ts.path, err)
	}
	tok := strings.TrimSpace(string(tokb))
	if len(tok) == 0 {
		return nil, fmt.Errorf("read empty token from file %q", ts.path)
	}

	return &oauth2.Token{
		AccessToken: tok,
		Expiry:      time.Now().Add(ts.period),
	}, nil
}

type cachingTokenSource struct {
	base   oauth2.TokenSource
	leeway time.Duration

	sync.RWMutex
	tok *oauth2.Token
	t   time.Time

	// for testing
	now func() time.Time
}

func (ts *cachingTokenSource) Token() (*oauth2.Token, error) {
	now := ts.now()
	// fast path
	ts.RLock()
	tok := ts.tok
	ts.RUnlock()

	if tok != nil && tok.Expiry.Add(-1*ts.leeway).After(now) {
		return tok, nil
	}

	// slow path
	ts.Lock()
	defer ts.Unlock()
	if tok := ts.tok; tok != nil && tok.Expiry.Add(-1*ts.leeway).After(now) {
		return tok, nil
	}

	tok, err := ts.base.Token()
	if err != nil {
		if ts.tok == nil {
			return nil, err
		}
		klog.Errorf("Unable to rotate token: %v", err)
		return ts.tok, nil
	}

	ts.t = ts.now()
	ts.tok = tok
	return tok, nil
}

func (ts *cachingTokenSource) ResetTokenOlderThan(t time.Time) {
	ts.Lock()
	defer ts.Unlock()
	if ts.t.Before(t) {
		ts.tok = nil
		ts.t = time.Time{}
	}
}
