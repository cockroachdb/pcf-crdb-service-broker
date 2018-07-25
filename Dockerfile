FROM golang:1.10-alpine AS build
RUN apk --no-cache add git make && \
  go get -u github.com/golang/dep/cmd/dep
WORKDIR $GOPATH/src/app
COPY . $GOPATH/src/app
RUN dep ensure && make build-static

FROM scratch
COPY --from=build /go/src/app/cockroachdb-servicebroker-static /cockroachdb-servicebroker-static
ENTRYPOINT ["/cockroachdb-servicebroker-static"]
