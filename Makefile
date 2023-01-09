.DEFAULT_GOAL := help
SHELL := /bin/bash

#help: @ list available tasks on this project
help:
	@grep -E '[a-zA-Z\.\-]+:.*?@ .*$$' $(MAKEFILE_LIST)| tr -d '#'  | awk 'BEGIN {FS = ":.*?@ "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

#test.unit: @ run unit tests and coverage
test.unit:
	@echo "[TEST.UNIT] run unit tests and coverage"
	@go test -timeout 30s -race -covermode=atomic -coverprofile=coverage.out github.com/SLedunois/bigbluebutton-telegraf-plugin/plugins/inputs/bigbluebutton

#build: @ build bigbluebutton telegraf plugin binary
build: 
	@echo "[BUILD] build bbsctl binary"
	rm -rf pkg
	./build.sh
