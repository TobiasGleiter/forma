## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## test: run tests with coverage
.PHONY: test
test:
	go test -shuffle=on -short -tags sqlite -vet=off -race -timeout 120s -covermode=atomic -coverprofile=/tmp/profile.out ./...

## coverage: 
.PHONY: coverage
coverage: test
	go tool cover -html=/tmp/profile.out

## lint: run linters
.PHONY: lint
lint:
	golangci-lint run --fix ./...