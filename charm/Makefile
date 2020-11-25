# clear default rules
MAKEFLAGS += --no-builtin-rules
ifndef CHARM_NAME
	$(error No CHARM_NAME defined)
endif

ifndef CHARM_PUBLISH_REPO
	$(error CHARM_PUBLISH_REPO is required)
endif

TMPDIR ?= $(CURDIR)/tmp_charm
CHARM_SRC ?= $(CURDIR)/charm
CHARM_SERIES ?= xenial
CHARM_EXCLUDE_FILE ?= build-exclude.txt
PROJECT_WHEEL_DIR ?= $(TMPDIR)/dependencies

BUILDDIR ?= $(CURDIR)/dist
CODETREE ?= /snap/bin/codetree

CHARM_REPO_DIR ?= $(BUILDDIR)/build

# for personal test builds
LP_USER = $(shell bzr launchpad-login)
PRIVATE_CHARM_REPO_DIR = $(BUILDDIR)/build-$(LP_USER)
PRIVATE_PUBLISH_REPO = $(shell echo "$(CHARM_PUBLISH_REPO)" | sed -e 's/~[^/]*\//~$(LP_USER)\//')

JUJU_REPOSITORY = $(BUILDDIR)
CHARM_BUILD_DIR = $(BUILDDIR)/$(CHARM_SERIES)/$(CHARM_NAME)


LAYER_PATH = $(TMPDIR)/ols-layers/layer
INTERFACE_PATH = $(TMPDIR)/ols-layers/interface
CHARM_WHEELS_DIR = $(TMPDIR)/charm-wheels
# Needed to build git email address
DOMAIN ?= $(shell hostname -f)
GIT_USERNAME = $(shell git config --get user.name || echo $(USER) )
GIT_EMAIL = $(shell git config --get user.email || echo $(USER)@$(DOMAIN) )


export INTERFACE_PATH
export LAYER_PATH
export JUJU_REPOSITORY

BASENODE_REPO ?= lp:~canonical-sysadmins/basenode/trunk
BASENODE_CACHE_DIR = $(TMPDIR)/basenode
BASENODE_CACHE = $(BASENODE_CACHE_DIR)/.done
BASENODE_DIR = $(CHARM_BUILD_DIR)/exec.d/
BASENODE = $(BASENODE_DIR)/.done

$(BUILDDIR) $(TMPDIR):
	mkdir -p $@

$(BASENODE_CACHE): | $(TMPDIR)
	bzr branch $(BASENODE_REPO) $(BASENODE_CACHE_DIR)
	touch $@

.PHONY: $(BASENODE)
$(BASENODE): | $(BASENODE_CACHE)
	mkdir -p $(BASENODE_DIR)
	cp -a $(BASENODE_CACHE_DIR) $(BASENODE_DIR)
	touch $@

$(CODETREE):
	sudo snap install codetree --classic

charm-dependencies: $(CHARM_DEPS)

CHARM_DEPS = $(LAYER_PATH)/.done $(INTERFACE_PATH)/.done $(CHARM_WHEELS_DIR)/.done
$(CHARM_DEPS): $(CHARM_SRC)/dependencies.txt $(CODETREE) | $(BASENODE_CACHE) $(TMPDIR)
	cd $(TMPDIR) && $(CODETREE) $<
	touch $(CHARM_DEPS)

.PHONY: $(CHARM_BUILD_DIR)
$(CHARM_BUILD_DIR): $(CHARM_DEPS) | $(BUILDDIR)
	rm -rf $(CHARM_BUILD_DIR)
	PIP_NO_INDEX=true PIP_FIND_LINKS=$(CHARM_WHEELS_DIR) charm build -o $(BUILDDIR) -s $(CHARM_SERIES) -n $(CHARM_NAME) ./charm/serial-vault

/snap/bin/charm:
	sudo snap install charm --classic

$(CHARM_REPO_DIR)/.git:
	git clone $(CHARM_PUBLISH_REPO) $(CHARM_REPO_DIR)

charm-build-repo: | $(CHARM_REPO_DIR)/.git
	cd $(CHARM_REPO_DIR) && git pull

charm-build: /snap/bin/charm $(CHARM_BUILD_DIR) $(BASENODE)

charm-collect: $(CHARM_DEPS)

charm-clean:
	rm -rf $(LAYER_PATH) $(INTERFACE_PATH) $(CHARM_WHEELS_DIR) $(TMPDIR)

charm-deploy: bundle.yaml
	@echo "Deploying $(CHARM_NAME)..."
	juju deploy ./charm/bundle.yaml

charm-upgrade:
	juju upgrade-charm $(CHARM_NAME) --path $(CHARM_BUILD_DIR)

charm-check: IP=$(shell juju status $(CHARM_NAME) --format=short |tail -1| awk '{print $$3}')
charm-check: PROXY=$(shell juju status haproxy --format=short |tail -1| awk '{print $$3}')
charm-check: PORT=$(shell grep port: charm/layer.yaml | awk '{print $$2}')
charm-check:
	@echo Checking $(CHARM_NAME) unit...
	curl http://$(IP):$(PORT)/_status/ping
	curl http://$(IP):$(PORT)/_status/check
	@echo Checking haproxy routing unit...
	curl http://$(PROXY)/_status/ping
	curl http://$(PROXY)/_status/check


charm-publish: TRUNK_REVNO=$(shell bzr revision-info)
charm-publish: charm-build charm-build-repo
	rsync -a -m --ignore-times --exclude .git --delete $(CHARM_BUILD_DIR)/ $(CHARM_REPO_DIR)/
	cd $(CHARM_REPO_DIR) && git add --no-ignore-removal .
	# Need to force username and email because git can't autodetect this
	# on jenkaas slaves.
	cd $(CHARM_REPO_DIR) && git -c user.name="$(GIT_USERNAME)" -c user.email="$(GIT_EMAIL)"  commit -am "Build of $(CHARM_NAME) from $(TRUNK_REVNO)"
	cd $(CHARM_REPO_DIR) && git tag "$(shell git rev-parse --short HEAD)"
	cd $(CHARM_REPO_DIR) && git push origin master --tags

charm-private-publish:
	git ls-remote -q $(PRIVATE_PUBLISH_REPO) || ( cd $$(mktemp -d); git init; git commit --allow-empty -m 'init'; git push $(PRIVATE_PUBLISH_REPO) master)
	$(MAKE) charm-publish CHARM_PUBLISH_REPO=$(PRIVATE_PUBLISH_REPO) CHARM_REPO_DIR=$(PRIVATE_CHARM_REPO_DIR)

# get swift credentials
-include ~/.config/swift/serial-vault
export $(shell sed 's/=.*//' ~/.config/swift/serial-vault)
export BUILD_LABEL=${VERSION}
export BUILD_CONTAINER_NAME=${SWIFT_CONTAINER_NAME}

bundle.yaml: ./charm/bundle.yaml.in
	envsubst < ./charm/bundle.yaml.in > ./charm/bundle.yaml
