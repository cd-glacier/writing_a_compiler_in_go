.PHONY: help
.DEFAULT_GOAL := help

deps: ## install development dependencies
	go get -u github.com/rakyll/gotest
	go get -u github.com/nathany/looper

test: ## test with gotest
	gotest -v ./src/...

watch-test: ## watch test
	cd src && looper

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
