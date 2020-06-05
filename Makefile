PROJECT_NAME := $(shell basename $(CURDIR))
.DEFAULT_GOAL := help

.PHONY:phony

fmt: phony ## format the codes
	@go fmt ./...

lint: phony ## lint the codes
	@golint ./...

vet: phony fmt ## vet the codes
	@go vet ./...

build: phony vet ## build the binary
	@go build

run: phony build ## run the binary
	@./$(PROJECT_NAME)

GREEN  := $(shell tput -Txterm setaf 2)
RESET  := $(shell tput -Txterm sgr0)

help: phony ## print this help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ { printf "${GREEN}%-20s${RESET}%s\n", $$1, $$NF }' $(MAKEFILE_LIST)
