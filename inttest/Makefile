.PHONY: all inttest inttest-csrx

.DEFAULT: inttest 

inttest: inttest-csrx

inttest-csrx:
	docker-compose -f docker-compose.csrx.yml up \
		--build \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from inttest