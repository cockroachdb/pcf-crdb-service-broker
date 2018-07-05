SOURCES=$(shell ls -1 *.go)
CRDBSB_IMAGE ?= cockroachdb-servicebroker
CRDBSB_TAG ?= 0.1

CRDB_HOST ?= cockroachdb-public.default.svc.cluster.local
CRDB_PORT ?= 26257
CRDB_TOKEN ?= configure-me-token

.PHONY: build
build: deps cockroachdb-servicebroker

.PHONY: build-static
build-static: deps cockroachdb-servicebroker-static

cockroachdb-servicebroker: $(SOURCES)
	go build -o $@ $(SOURCES)

cockroachdb-servicebroker-static: $(SOURCES)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $@ $(SOURCES)

.PHONY: docker-build
docker-build:
	docker build -t $(CRDBSB_IMAGE):$(CRDBSB_TAG) .

.PHONY: docker-push
docker-push: docker-build
	docker push $(CRDBSB_IMAGE):$(CRDBSB_TAG)

.PHONY: deps
deps:
	dep ensure

.PHONY: deploy
.EXPORT_ALL_VARIABLES: deploy
deploy:
	# $DOLLAR enables environment variables i.e. ${DOLLAR}${VAR}
	DOLLAR='$$' envsubst < kubernetes/servicebroker.yaml | kubectl apply -f -
