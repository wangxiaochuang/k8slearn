package options

import (
	"fmt"
	"time"

	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
)

type SecureServingOptionsWithLoopback struct {
	*SecureServingOptions
}

func (o *SecureServingOptions) WithLoopback() *SecureServingOptionsWithLoopback {
	return &SecureServingOptionsWithLoopback{o}
}

func (s *SecureServingOptionsWithLoopback) ApplyTo(secureServingInfo **server.SecureServingInfo, loopbackClientConfig **rest.Config) error {
	if s == nil || s.SecureServingOptions == nil || secureServingInfo == nil {
		return nil
	}

	if err := s.SecureServingOptions.ApplyTo(secureServingInfo); err != nil {
		return err
	}

	if *secureServingInfo == nil || loopbackClientConfig == nil {
		return nil
	}

	// 为本地调用的连接创建自签名证书
	certPem, keyPem, err := certutil.GenerateSelfSignedCertKey(server.LoopbackClientServerNameOverride, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to generate self-signed certificate for loopback connection: %v", err)
	}

	time.Sleep(time.Hour)

	panic("not implemented")
}
