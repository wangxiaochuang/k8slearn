#!/bin/bash

token="$1"

run() {
    param="limit=500&resourceVersion=0"
    if [ "$2" != "" ]; then
        param=$2
    fi
    curl -ki -H "Origin: https://127.0.0.1:6379" \
        -H "Authorization: Bearer ${token}" \
        -H "User-Agent: kube-apiserver/v0.0.0 (darwin/amd64) kubernetes/4ce5a89" \
        "https://127.0.0.1:6443$1?wxcdebug=true&${param}"
}

#run "/api/v1/nodes"
run "/api/v1/namespaces/default"
