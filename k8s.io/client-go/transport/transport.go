package transport

import "net/http"

type WrapperFunc func(rt http.RoundTripper) http.RoundTripper
