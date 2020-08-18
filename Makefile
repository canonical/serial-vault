BASEDIR=${CURDIR}
VENDOR=${BASEDIR}/vendor
TMP=${BASEDIR}/tmp
VENDOR_TMP=${TMP}/vendor
LOCAL_BIN:=${TMP}/bin
GOBIN=${BASEDIR}/bin

GIT_REVISION = $(shell git rev-parse --short HEAD)
VERSION      ?= $(shell git describe --tags --abbrev=0)
JOBDATE		 ?= $(shell date -u +%Y-%m-%dT%H%M%SZ)

# --ldflags sets the flags that are passed to 'go tool link'
LDFLAGS += -X github.com/CanonicalLtd/serial-vault/config.version=$(VERSION)
LDFLAGS += -X github.com/CanonicalLtd/serial-vault/config.revision=$(GIT_REVISION)
# This will make the linked C part also static into the binary (can produce some warnings)
LDFLAGS_STATIC += $(LDFLAGS) -linkmode external -extldflags -static

GOFLAGS=-mod=vendor
# make sure we use built-in net package and not the system’s one
GOTAGS=-tags netgo

SERVICE_NAME=serial-vault
LOCAL_SERVICE_NAME = ${GOBIN}/${SERVICE_NAME}
GO ?= go

# this repo contains external dependencies for an internal build
# VENDOR_BRANCH_URL ?= lp:~ubuntuone-pqm-team/serial-vault/+git/dependencies
VENDOR_BRANCH_URL ?= lp:~glower/serial-vault/+git/dependencies

default: build-sv

# run build-in database migration
migrate:
	$(GO) run cmd/serial-vault-admin/main.go database --config=settings.yaml

bootstrap: vendor mkdir-tmp

mkdir-tmp:
	mkdir -p $(TMP)

install-static:
	$(info # Installing binaries into $(GOBIN))
	@GOBIN=$(GOBIN) $(GO) install $(GOFLAGS) $(GOTAGS) -ldflags "$(LDFLAGS_STATIC) -w" -v ./...

install:
	$(info # Installing binaries into $(GOBIN))
	@GOBIN=$(GOBIN) $(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS) -w" -v ./...

build-static:
	$(info # Building ${SERVICE_NAME} binaries)
	@cd cmd/serial-vault && $(GO) build -a $(GOFLAGS) $(GOTAGS) -ldflags "$(LDFLAGS_STATIC) -w" -o $(LOCAL_SERVICE_NAME)

build:
	$(info # Building ${SERVICE_NAME} binaries)
	@cd cmd/serial-vault && $(GO) build -a $(GOFLAGS) -ldflags "$(LDFLAGS) -w" -o $(LOCAL_SERVICE_NAME)

run: run-admin

run-admin: build
	$(info # Running ${SERVICE_NAME} in admin/ui mode)
	@CSRF_SECURE=disable ${LOCAL_SERVICE_NAME} --mode=admin --config=settings.yaml

run-sign: build
	$(info # Running ${SERVICE_NAME} in sign/api mode)
	@${LOCAL_SERVICE_NAME} --mode=sign --config=settings.yaml

# get the vendor code for internal build
vendor:
	[ -d $(VENDOR) ] && (cd $(VENDOR) && git pull) || (git clone $(VENDOR_BRANCH_URL) $(VENDOR))

# if you need to add an additional external dependency, use this target
vendoring: mkdir-tmp
	@rm -rf ${VENDOR_TMP}
	@${GO} mod vendor
	@${GO} mod tidy
	@mv ${VENDOR} ${VENDOR_TMP}
	@git clone $(VENDOR_BRANCH_URL) $(VENDOR)
	@cp -r ${VENDOR_TMP} .
	@cd ${VENDOR} && git add . && git checkout -b vendoring-$(JOBDATE)
	@echo "\n!!! Please go to $(VENDOR) folder check the changes and create a MP !!!\n"

unit-test:
	$(info # Running unit tests for ${SERVICE_NAME})
	./run-checks --unit

static-test:
	$(info # Running static checks for ${SERVICE_NAME})
	@go get -u golang.org/x/lint/golint
	./run-checks --static

test: unit-test static-test

# only for dev/testing, don't commit output of this command by yourself
# it will be done automatically by CI 
build-frontend:
	cd webapp-admin && \
	npm install && \
	npm run build && \
	rm ../static/js/* && \
	rm ../static/css/* && \
	cp -R build/static/js ../static && \
	cp -R build/static/css ../static && \
	cp -R build/index.html ../static/app.html

# run application/db in docker
run-docker:
	cd docker-compose && docker-compose up

clean:
	rm -rf ${TMP}
	rm -rf ${GOBIN}/factory
	rm -rf ${GOBIN}/serial-vault
	rm -rf ${GOBIN}/serial-vault-admin

.PHONY: bootstrap vendor run clean install build test migrate
