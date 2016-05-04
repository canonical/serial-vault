FROM golang:1.6.2

ADD . /go/src/github.com/ubuntu-core/identity-vault

WORKDIR /go/src/github.com/ubuntu-core/identity-vault
# get dependency
RUN go get ./...

COPY ./docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

