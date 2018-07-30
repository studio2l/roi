#!/bin/bash

if [ -z $GOROOT ]; then
	GOROOT="/usr/local/go"
fi

"$GOROOT/bin/go" run "$GOROOT/src/crypto/tls/generate_cert.go" -host="localhost" -ca=true
