#!/bin/bash

if [ "`ps -ef |grep etcd |grep -v 'grep'`" = "" ]; then
    cd /opt/etcd
    nohup etcd &
    cd -
fi

go run k8s.io/kubernetes/cmd/kube-apiserver $* \
    --cert-dir=/tmp/kubernetes \
    --service-account-issuer=xxxxx \
    --etcd-servers=http://127.0.0.1:2379 \
    --egress-selector-config-file=testdata/egress-selector.yml \
    --feature-gates="APIServerTracing=true" \
    --tracing-config-file=testdata/tracing-config.yml
