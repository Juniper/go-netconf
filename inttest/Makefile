.ONESHELL:
.PHONY: all inttest inttest-csrx

.DEFAULT: help 

help:
	@echo "available targets"
	@echo "    csrx  Run integration against csrx container (\$$CSRX_IMAGE:$(CSRX_IMAGE))"
	@echo "    ceos  Run integration against ceos container (\$$CEOS_IMAGE:$(CEOS_IMAGE))" 
	@echo "    all   Run all integration tests"

all: csrx ceos

CSRX_IMAGE ?= "csrx:20.3R1.8"
csrx:
	CSRX_IMAGE=$(CSRX_IMAGE) docker-compose -f docker-compose.csrx.yml up \
		--build \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from inttest

CEOS_IMAGE ?= "ceos64-lab:4.28.3M"
ceos:
	CEOS_IMAGE=$(CEOS_IMAGE) docker-compose -f docker-compose.ceos.yml up \
		--build \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from inttest

CONFD_IMAGE ?= "confd-basic:7.8.3"
confd:
	CONFD_IMAGE=$(CONFD_IMAGE) docker-compose -f docker-compose.confd.yml up \
		--build \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from inttest

netopeer2:
	docker-compose -f docker-compose.netopeer2.yml up \
		--build \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from inttest
   