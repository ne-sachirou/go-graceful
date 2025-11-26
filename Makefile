SHELL=/bin/bash

.PHONY: help
help:
	@awk -F':.*##' '/^[-_a-zA-Z0-9]+:.*##/{printf"%-12s\t%s\n",$$1,$$2}' $(MAKEFILE_LIST) | sort

lint: lint-gha lint-go lint-markdown lint-renovate ## Lint
	yamllint .

lint-gha:
	yamllint .github/workflows/
	actionlint
	ghalint run
	zizmor .

lint-go:
	go vet || true

lint-markdown:
	npx prettier -c README.md

lint-renovate:
	npx --package renovate -- renovate-config-validator

