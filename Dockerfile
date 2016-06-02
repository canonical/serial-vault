FROM golang:1.6.2

RUN apt-get update && apt-get install -y postgresql-client
ADD . /go/src/github.com/ubuntu-core/identity-vault

WORKDIR /go/src/github.com/ubuntu-core/identity-vault
# get dependency
RUN sh -c "go get launchpad.net/godeps; godeps -t -u dependencies.tsv"

COPY ./docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

