#!/bin/bash
set -e

CERT_DIR=$(dirname $BASH_SOURCE)/db-cert

if [ -f $CERT_DIR/ca.key ]; then
	echo "db already initialized."
	exit 1
fi

cockroach cert create-ca --certs-dir=$CERT_DIR --ca-key=$CERT_DIR/ca.key
cockroach cert create-node --certs-dir=$CERT_DIR --ca-key=$CERT_DIR/ca.key localhost
cockroach cert create-client --certs-dir=$CERT_DIR --ca-key=$CERT_DIR/ca.key root

