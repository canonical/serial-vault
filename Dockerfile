FROM ubuntu:xenial

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y postgresql-client golang-1.13-go ca-certificates git

ENV GOPATH=/go
ADD . /go/src/github.com/CanonicalLtd/serial-vault

WORKDIR /go/src/github.com/CanonicalLtd/serial-vault
# get dependencies
RUN /usr/lib/go-1.13/bin/go get ./...

COPY ./docker-compose/settings.yaml /go/src/github.com/CanonicalLtd/serial-vault
COPY ./docker-compose/docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]
