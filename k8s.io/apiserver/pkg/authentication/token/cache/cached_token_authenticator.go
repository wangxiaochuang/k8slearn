package cache

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sync/singleflight"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/utils/clock"
)

var errAuthnCrash = apierrors.NewInternalError(errors.New("authentication failed unexpectedly"))

const sharedLookupTimeout = 30 * time.Second

type cacheRecord struct {
	resp        *authenticator.Response
	ok          bool
	err         error
	annotations map[string]string
}

type cachedTokenAuthenticator struct {
	authenticator authenticator.Token

	cacheErrs  bool
	successTTL time.Duration
	failureTTL time.Duration

	cache cache
	group singleflight.Group

	// hashPool is a per authenticator pool of hash.Hash (to avoid allocations from building the Hash)
	// HMAC with SHA-256 and a random key is used to prevent precomputation and length extension attacks
	// It also mitigates hash map DOS attacks via collisions (the inputs are supplied by untrusted users)
	hashPool *sync.Pool
}

type cache interface {
	// given a key, return the record, and whether or not it existed
	get(key string) (value *cacheRecord, exists bool)
	// caches the record for the key
	set(key string, value *cacheRecord, ttl time.Duration)
	// removes the record for the key
	remove(key string)
}

func New(authenticator authenticator.Token, cacheErrs bool, successTTL, failureTTL time.Duration) authenticator.Token {
	return newWithClock(authenticator, cacheErrs, successTTL, failureTTL, clock.RealClock{})
}

func newWithClock(authenticator authenticator.Token, cacheErrs bool, successTTL, failureTTL time.Duration, clock clock.Clock) authenticator.Token {
	randomCacheKey := make([]byte, 32)
	if _, err := rand.Read(randomCacheKey); err != nil {
		panic(err) // rand should never fail
	}

	return &cachedTokenAuthenticator{
		authenticator: authenticator,
		cacheErrs:     cacheErrs,
		successTTL:    successTTL,
		failureTTL:    failureTTL,
		cache:         newStripedCache(32, fnvHashFunc, func() cache { return newSimpleCache(clock) }),

		hashPool: &sync.Pool{
			New: func() interface{} {
				return hmac.New(sha256.New, randomCacheKey)
			},
		},
	}
}

func (a *cachedTokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}

func keyFunc(hashPool *sync.Pool, auds []string, token string) string {
	h := hashPool.Get().(hash.Hash)

	h.Reset()

	// try to force stack allocation
	var a [4]byte
	b := a[:]

	writeLengthPrefixedString(h, b, token)
	// encode the length of audiences to avoid ambiguities
	writeLength(h, b, len(auds))
	for _, aud := range auds {
		writeLengthPrefixedString(h, b, aud)
	}

	key := toString(h.Sum(nil)) // skip base64 encoding to save an allocation

	hashPool.Put(h)

	return key
}

func writeLengthPrefixedString(w io.Writer, b []byte, s string) {
	writeLength(w, b, len(s))
	if _, err := w.Write(toBytes(s)); err != nil {
		panic(err) // Write() on hash never fails
	}
}

func writeLength(w io.Writer, b []byte, length int) {
	binary.BigEndian.PutUint32(b, uint32(length))
	if _, err := w.Write(b); err != nil {
		panic(err) // Write() on hash never fails
	}
}

func toBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// toString performs unholy acts to avoid allocations
func toString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
