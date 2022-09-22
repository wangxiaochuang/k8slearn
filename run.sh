#!/bin/bash

go run k8s.io/kubernetes/cmd/kube-apiserver $* \
    --cert-dir=/tmp/kubernetes \
    --service-account-issuer=xxxxx \
    --etcd-servers=http://127.0.0.1:2379 \
    --egress-selector-config-file=testdata/egress-selector.yml
