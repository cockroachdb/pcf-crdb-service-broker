#!/bin/sh

set -e

mkdir -p build
zip build/crdb-service-broker.zip -r . -x *.git* product/\* release/\* examples/\*

tile build
