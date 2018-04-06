FROM golang:1.9.4

RUN apt-get update && apt-get install -y postgresql-client
ADD . /go/src/github.com/CanonicalLtd/serial-vault

WORKDIR /go/src/github.com/CanonicalLtd/serial-vault
# get dependency
RUN sh -c "go get launchpad.net/godeps; godeps -t -u dependencies.tsv"

COPY ./docker-compose/settings.yaml /go/src/github.com/CanonicalLtd/serial-vault
COPY ./docker-compose/docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]
