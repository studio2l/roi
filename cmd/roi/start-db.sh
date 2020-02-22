#!/bin/bash
set -e

CERT_DIR=$(dirname $BASH_SOURCE)/db-cert

cockroach start-single-node --certs-dir=$CERT_DIR --listen-addr=localhost:26257
