.ONESHELL:warning
.PHONY: all inttest inttest-csrx

.DEFAULT: inttest 

inttest: inttest-csrx

CSRX_IMAGE ?= "csrx:20.3R1.8"
inttest-csrx:
	CSRX_IMAGE=$(CSRX_IMAGE) docker-compose -f docker-compose.csrx.yml up \
		--build \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from inttest