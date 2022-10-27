package serviceaccount

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	apiserverserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
)

type ServiceAccountTokenGetter interface {
	GetServiceAccount(namespace, name string) (*v1.ServiceAccount, error)
	GetPod(namespace, name string) (*v1.Pod, error)
	GetSecret(namespace, name string) (*v1.Secret, error)
}

type TokenGenerator interface {
	GenerateToken(claims *jwt.Claims, privateClaims interface{}) (string, error)
}

func JWTTokenGenerator(iss string, privateKey interface{}) (TokenGenerator, error) {
	var signer jose.Signer
	var err error
	switch pk := privateKey.(type) {
	case *rsa.PrivateKey:
		signer, err = signerFromRSAPrivateKey(pk)
		if err != nil {
			return nil, fmt.Errorf("could not generate signer for RSA keypair: %v", err)
		}
	case *ecdsa.PrivateKey:
		signer, err = signerFromECDSAPrivateKey(pk)
		if err != nil {
			return nil, fmt.Errorf("could not generate signer for ECDSA keypair: %v", err)
		}
	case jose.OpaqueSigner:
		signer, err = signerFromOpaqueSigner(pk)
		if err != nil {
			return nil, fmt.Errorf("could not generate signer for OpaqueSigner: %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown private key type %T, must be *rsa.PrivateKey, *ecdsa.PrivateKey, or jose.OpaqueSigner", privateKey)
	}

	return &jwtTokenGenerator{
		iss:    iss,
		signer: signer,
	}, nil
}

func keyIDFromPublicKey(publicKey interface{}) (string, error) {
	publicKeyDERBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to serialize public key to DER format: %v", err)
	}

	hasher := crypto.SHA256.New()
	hasher.Write(publicKeyDERBytes)
	publicKeyDERHash := hasher.Sum(nil)

	keyID := base64.RawURLEncoding.EncodeToString(publicKeyDERHash)

	return keyID, nil
}

func signerFromRSAPrivateKey(keyPair *rsa.PrivateKey) (jose.Signer, error) {
	keyID, err := keyIDFromPublicKey(&keyPair.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive keyID: %v", err)
	}

	privateJWK := &jose.JSONWebKey{
		Algorithm: string(jose.RS256),
		Key:       keyPair,
		KeyID:     keyID,
		Use:       "sig",
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key:       privateJWK,
		},
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %v", err)
	}

	return signer, nil
}

func signerFromECDSAPrivateKey(keyPair *ecdsa.PrivateKey) (jose.Signer, error) {
	var alg jose.SignatureAlgorithm
	switch keyPair.Curve {
	case elliptic.P256():
		alg = jose.ES256
	case elliptic.P384():
		alg = jose.ES384
	case elliptic.P521():
		alg = jose.ES512
	default:
		return nil, fmt.Errorf("unknown private key curve, must be 256, 384, or 521")
	}

	keyID, err := keyIDFromPublicKey(&keyPair.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive keyID: %v", err)
	}

	// Wrap the ECDSA keypair in a JOSE JWK with the designated key ID.
	privateJWK := &jose.JSONWebKey{
		Algorithm: string(alg),
		Key:       keyPair,
		KeyID:     keyID,
		Use:       "sig",
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: alg,
			Key:       privateJWK,
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %v", err)
	}

	return signer, nil
}

func signerFromOpaqueSigner(opaqueSigner jose.OpaqueSigner) (jose.Signer, error) {
	alg := jose.SignatureAlgorithm(opaqueSigner.Public().Algorithm)

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: alg,
			Key: &jose.JSONWebKey{
				Algorithm: string(alg),
				Key:       opaqueSigner,
				KeyID:     opaqueSigner.Public().KeyID,
				Use:       "sig",
			},
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %v", err)
	}

	return signer, nil
}

type jwtTokenGenerator struct {
	iss    string
	signer jose.Signer
}

func (j *jwtTokenGenerator) GenerateToken(claims *jwt.Claims, privateClaims interface{}) (string, error) {
	// claims are applied in reverse precedence
	return jwt.Signed(j.signer).
		Claims(privateClaims).
		Claims(claims).
		Claims(&jwt.Claims{
			Issuer: j.iss,
		}).
		CompactSerialize()
}

// p227
func JWTTokenAuthenticator(issuers []string, keys []interface{}, implicitAuds authenticator.Audiences, validator Validator) authenticator.Token {
	issuersMap := make(map[string]bool)
	for _, issuer := range issuers {
		issuersMap[issuer] = true
	}
	return &jwtTokenAuthenticator{
		issuers:      issuersMap,
		keys:         keys,
		implicitAuds: implicitAuds,
		validator:    validator,
	}
}

type jwtTokenAuthenticator struct {
	issuers      map[string]bool
	keys         []interface{}
	validator    Validator
	implicitAuds authenticator.Audiences
}

type Validator interface {
	Validate(ctx context.Context, tokenData string, public *jwt.Claims, private interface{}) (*apiserverserviceaccount.ServiceAccountInfo, error)
	NewPrivateClaims() interface{}
}

func (j *jwtTokenAuthenticator) AuthenticateToken(ctx context.Context, tokenData string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
