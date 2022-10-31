#!/bin/bash

if [ "`ps -ef |grep etcd |grep -v 'grep'`" = "" ]; then
    cd /opt/etcd
    nohup etcd &
    cd -
fi

if [ "$1" = "test" ]; then
    go run k8s.io/kubernetes/cmd/mytest $* \
    exit 0
fi

go run k8s.io/kubernetes/cmd/kube-apiserver $* \
    --authorization-mode=Node,RBAC \
    --cloud-provider= --cloud-config= --vmodule= \
    --v=0 --audit-policy-file=testdata/kube-audit-policy-file \
    --audit-log-path=/tmp/kube-apiserver-audit.log \
    --authorization-webhook-config-file= --authentication-token-webhook-config-file= \
    --cert-dir=./cert-dir/ \
    --egress-selector-config-file=testdata/kube_egress_selector_configuration.yaml \
    --client-ca-file=cert-dir/client-ca.crt \
    --kubelet-client-certificate=cert-dir/client-kube-apiserver.crt \
    --kubelet-client-key=cert-dir/client-kube-apiserver.key \
    --service-account-key-file=cert-dir/kube-serviceaccount.key \
    --service-account-lookup=true \
    --service-account-issuer=https://kubernetes.default.svc \
    --service-account-jwks-uri=https://kubernetes.default.svc/openid/v1/jwks \
    --service-account-signing-key-file=cert-dir/kube-serviceaccount.key \
    --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,Priority,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,NodeRestriction \
    --disable-admission-plugins= --admission-control-config-file= \
    --bind-address=0.0.0.0 --secure-port=6443 \
    --tls-cert-file=cert-dir/serving-kube-apiserver.crt \
    --tls-private-key-file=cert-dir/serving-kube-apiserver.key \
    --storage-backend=etcd3 \
    --storage-media-type=application/vnd.kubernetes.protobuf \
    --etcd-servers=http://127.0.0.1:2379 \
    --service-cluster-ip-range=10.0.0.0/24 \
    --feature-gates=AllAlpha=false \
    --external-hostname=localhost \
    --requestheader-username-headers=X-Remote-User \
    --requestheader-group-headers=X-Remote-Group \
    --requestheader-extra-headers-prefix=X-Remote-Extra- \
    --requestheader-client-ca-file=cert-dir/request-header-ca.crt \
    --requestheader-allowed-names=system:auth-proxy \
    --proxy-client-cert-file=cert-dir/client-auth-proxy.crt \
    --proxy-client-key-file=cert-dir/client-auth-proxy.key \
    --cors-allowed-origins='/127.0.0.1(:[0-9]+)?$,/localhost(:[0-9]+)?$'
