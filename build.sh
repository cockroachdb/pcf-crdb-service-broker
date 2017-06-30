#!/bin/sh

zip build/crdb-service-broker.zip -r . -x *.git* product/\* release/\* examples/\*

tile build
