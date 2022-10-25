package dynamiccertificates

import (
	"bytes"
	"context"
	"crypto/x509"
	"strings"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type unionCAContent []CAContentProvider

var _ CAContentProvider = &unionCAContent{}
var _ ControllerRunner = &unionCAContent{}

func NewUnionCAContentProvider(caContentProviders ...CAContentProvider) CAContentProvider {
	return unionCAContent(caContentProviders)
}

func (c unionCAContent) Name() string {
	names := []string{}
	for _, curr := range c {
		names = append(names, curr.Name())
	}
	return strings.Join(names, ",")
}

func (c unionCAContent) CurrentCABundleContent() []byte {
	caBundles := [][]byte{}
	for _, curr := range c {
		if currCABytes := curr.CurrentCABundleContent(); len(currCABytes) > 0 {
			caBundles = append(caBundles, []byte(strings.TrimSpace(string(currCABytes))))
		}
	}

	return bytes.Join(caBundles, []byte("\n"))
}

func (c unionCAContent) VerifyOptions() (x509.VerifyOptions, bool) {
	currCABundle := c.CurrentCABundleContent()
	if len(currCABundle) == 0 {
		return x509.VerifyOptions{}, false
	}

	// TODO make more efficient.  This isn't actually used in any of our mainline paths.  It's called to build the TLSConfig
	// TODO on file changes, but the actual authentication runs against the individual items, not the union.
	ret, err := newCABundleAndVerifier(c.Name(), c.CurrentCABundleContent())
	if err != nil {
		// because we're made up of already vetted values, this indicates some kind of coding error
		panic(err)
	}

	return ret.verifyOptions, true
}

func (c unionCAContent) AddListener(listener Listener) {
	for _, curr := range c {
		curr.AddListener(listener)
	}
}

func (c unionCAContent) RunOnce(ctx context.Context) error {
	errors := []error{}
	for _, curr := range c {
		if controller, ok := curr.(ControllerRunner); ok {
			if err := controller.RunOnce(ctx); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return utilerrors.NewAggregate(errors)
}

func (c unionCAContent) Run(ctx context.Context, workers int) {
	for _, curr := range c {
		if controller, ok := curr.(ControllerRunner); ok {
			go controller.Run(ctx, workers)
		}
	}
}
