#!/bin/bash
set -e

CERT_DIR=$(dirname $BASH_SOURCE)/db-cert

ADDR=localhost:26257
if [ ! -z $ROI_DB_ADDR ]; then
	ADDR=$ROI_DB_ADDR
fi

HTTP_ADDR=localhost:8080
if [ ! -z $ROI_DB_HTTP_ADDR ]; then
	HTTP_ADDR=$ROI_DB_HTTP_ADDR
fi

cockroach start-single-node --certs-dir=$CERT_DIR --http-addr=$HTTP_ADDR --listen-addr=$ADDR
