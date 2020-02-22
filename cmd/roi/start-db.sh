#!/bin/bash
set -e

CERT_DIR=$(dirname $BASH_SOURCE)/db-cert

ADDR=localhost:26257
if [ ! -z $ROI_DB_ADDR ]; then
	ADDR=$ROI_DB_ADDR
fi
cockroach start-single-node --certs-dir=$CERT_DIR --listen-addr=$ADDR
