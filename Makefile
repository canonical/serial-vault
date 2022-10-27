BASEDIR=${CURDIR}
VENDOR=${BASEDIR}/vendor
TMP=${BASEDIR}/tmp
VENDOR_TMP=${TMP}/vendor
LOCAL_BIN:=${TMP}/bin
GOBIN=${BASEDIR}/bin

GIT_REVISION = $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --abbrev=0)
JOBDATE ?= $(shell date -u +%Y-%m-%dT%H%M%SZ)

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
TAR_BASE_NAME = ${SERVICE_NAME}-${VERSION}.tar
TGZ_NAME = ${TMP}/${TAR_BASE_NAME}.gz
SWIFT_CONTAINER_NAME = ${SERVICE_NAME}.canonical.com

GO ?= go

# this repo contains external dependencies for an internal build
VENDOR_BRANCH_URL ?= lp:~ubuntuone-pqm-team/serial-vault/+git/dependencies

.PHONY: default
default: build

# run build-in database migration
.PHONY: migrate
migrate:
	$(GO) run cmd/serial-vault-admin/main.go database --config=settings.yaml

.PHONY: bootstrap
bootstrap: vendor mkdir-tmp

.PHONY: mkdir-tmp
mkdir-tmp:
	mkdir -p $(TMP)

.PHONY: install-static
install-static:
	$(info # Installing binaries into $(GOBIN))
	GOBIN=$(GOBIN) $(GO) install $(GOFLAGS) $(GOTAGS) -ldflags "$(LDFLAGS_STATIC) -w" -v ./...


.PHONY: install
install:
	$(info # Installing binaries into $(GOBIN))
	# remove the entire module download and build cache
	$(GO) clean -cache -modcache
	GOBIN=$(GOBIN) $(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS) -w" -v github.com/CanonicalLtd/serial-vault/cmd/serial-vault
	GOBIN=$(GOBIN) $(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS) -w" -v github.com/CanonicalLtd/serial-vault/cmd/serial-vault-admin
	GOBIN=$(GOBIN) $(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS) -w" -v github.com/CanonicalLtd/serial-vault/cmd/factory

.PHONY: build-static
build-static:
	$(info # Building ${SERVICE_NAME} binaries)
	cd cmd/serial-vault && $(GO) build -a $(GOFLAGS) $(GOTAGS) -ldflags "$(LDFLAGS_STATIC) -w" -o $(LOCAL_SERVICE_NAME)

.PHONY: build
build:
	$(info # Building ${SERVICE_NAME} binaries)
	cd cmd/serial-vault && $(GO) build -a $(GOFLAGS) -ldflags "$(LDFLAGS) -w" -o $(LOCAL_SERVICE_NAME)

.PHONY: run
run: run-admin

.PHONY: run-admin
run-admin: build
	$(info # Running ${SERVICE_NAME} in admin/ui mode)
	CSRF_SECURE=disable ${LOCAL_SERVICE_NAME} --mode=admin --config=settings.yaml

.PHONY: run-sign
run-sign: build
	$(info # Running ${SERVICE_NAME} in sign/api mode)
	${LOCAL_SERVICE_NAME} --mode=sign --config=settings.yaml

# get the vendor code for internal CI build
.PHONY: vendor-ci
vendor-ci:
	[ -d $(VENDOR) ] && (cd $(VENDOR) && git pull) || (git clone $(VENDOR_BRANCH_URL) $(VENDOR))

# get the vendor code
.PHONY: vendor
vendor:
	$(GO) mod vendor

# if you need to add an additional external dependency, use this target
.PHONY: vendoring-ci
vendoring-ci: mkdir-tmp
	rm -rf ${VENDOR_TMP}
	${GO} mod vendor
	${GO} mod tidy
	mv ${VENDOR} ${VENDOR_TMP}
	git clone $(VENDOR_BRANCH_URL) $(VENDOR)
	cp -r ${VENDOR_TMP} .
	cd ${VENDOR} && git add . && git checkout -b vendoring-$(JOBDATE)
	@echo "\n!!! Please go to $(VENDOR) folder check the changes and create a MP !!!\n"

.PHONY: unit-test
unit-test:
	$(info # Running unit tests for ${SERVICE_NAME})
	./run-checks --unit

.PHONY: static-test
static-test:
	$(info # Running static checks for ${SERVICE_NAME})
	GOBIN=$(GOBIN) ${GO} get -u golang.org/x/lint/golint
	GOBIN=$(GOBIN) ${GO} get -u github.com/client9/misspell/cmd/misspell
	GOBIN=$(GOBIN) ${GO} get -u github.com/gordonklaus/ineffassign
	GOBIN=$(GOBIN) ${GO} get -u github.com/alexkohler/nakedret
	./run-checks --static

.PHONY: test
test: unit-test static-test

.PHONY: build-tarball
build-tarball: install
	$(info # Creating tarball ${TGZ_NAME} with binaries and assets...)
	# create tar file with  assets in 'static' folder
	tar -cvf ${TAR_BASE_NAME} static
	# update tar file with serial-vault and serial-vault-admin built binaries
	tar -uvf ${TAR_BASE_NAME} -C ${GOBIN} serial-vault serial-vault-admin
	# compress with gzip and send directly to .tgz file including the release
	gzip -9 -c ${TAR_BASE_NAME} > ${TGZ_NAME}
	# calculate md5sum and create file with its value
	md5sum ${TGZ_NAME} > ${TGZ_NAME}.md5

SWIFT_TARGET_NAME=serial-vault-builds/$(VERSION)/serial-vault.tar.gz
.PHONY: publish-tarball
publish-tarball: vendor-ci mkdir-tmp build-tarball
	[ ! -e ~/.config/swift/serial-vault ] || . ~/.config/swift/serial-vault; \
	./publish-to-swift --debug $(SWIFT_CONTAINER_NAME) $(SWIFT_TARGET_NAME) $(TGZ_NAME) serial-vault=$(VERSION)
	./publish-to-swift --debug $(SWIFT_CONTAINER_NAME) ${SWIFT_TARGET_NAME}.md5 $(TGZ_NAME).md5 serial-vault=$(VERSION)

# only for dev/testing, don't commit output of this command by yourself
# it will be done automatically by CI 
.PHONY: build-frontend
build-frontend:
	cd webapp-admin && \
	npm install && \
	npm run build && \
	rm ../static/js/* && \
	rm ../static/css/* && \
	cp -R build/static/js ../static && \
	cp -R build/static/css ../static && \
	cp -R build/index.html ../static/app.html

.PHONY: test-frontend
test-frontend:
	NODE_ENV=test cd webapp-admin && \
	npm install && \
	npm install -g codecov && \
	npm run test

.PHONY: test-frontend-ci
test-frontend-ci:
	@echo "run frontend tests"
	NODE_ENV=test cd webapp-admin && \
	npm install && \
	npm install -g codecov && \
	npm run test:ci

# run application/db in docker
.PHONY: run-docker
run-docker:
	cd docker-compose && docker-compose up

# stop and remove containers
.PHONY:stop-docker
stop-docker:
	cd docker-compose && docker-compose kill && docker-compose rm

# essential charm makefile config
CHARM_NAME = ${SERVICE_NAME}
CHARM_PUBLISH_REPO ?= lp:~ubuntuone-pqm-team/serial-vault/+git/charm-build

# get charm build dependencies
.PHONY: charm-bootstrap
charm-bootstrap: $(CHARM_DEPS) vendor-ci

.PHONY: clean
clean:
	rm -rf ${TMP}
	rm -rf ${GOBIN}/factory
	rm -rf ${GOBIN}/serial-vault
	rm -rf ${GOBIN}/serial-vault-admin

include charm/Makefile
