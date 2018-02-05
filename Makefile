SOURCE_VERSION = $(shell git describe --tags --always --dirty --abbrev=6)
EXECUTABLE = big-iot-gateway
BUILD_DIR = build
GOBUILD = go build
PACKAGE = .
SRC_TARGET = ./cmd/server
OK_CMD = echo "\033[92mDone! "
ERR_CMD = echo "\033[91mError! "

UNAME := $(shell uname)
ifeq ($(UNAME), Linux)
	BUILDFLAGS = -v -x -ldflags "-linkmode external -extldflags -static"
endif
ifeq ($(UNAME), Darwin)
	BUILDFLAGS = -v -x
endif

default: help

.PHONY: install
install:
	dep ensure -v

.PHONY: build
build: ## Build a linux x64 executable
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILDFLAGS) -o $(BUILD_DIR)/$(EXECUTABLE) $(SRC_TARGET) && $(OK_CMD) || $(ERR_CMD)

.PHONY: build-darwin
build-darwin: ## Build a darwin x64 executable
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILDFLAGS) -o $(BUILD_DIR)/$(EXECUTABLE)-darwin $(SRC_TARGET) && $(OK_CMD) || $(ERR_CMD)

.PHONY: run 
run: ## Run binary on docker container
	docker-compose build --force-rm
	docker-compose run --service-ports big_iot_gw

.PHONY: release
release: ## Create target container with the release version of the app
	docker build --force-rm -t thingful/pomelo:$(SOURCE_VERSION) -t thingful/big-iot-gateway:$(SOURCE_VERSION) .

.PHONY: deploy-heroku
deploy-heroku: ## Build Container and deploys it into heroku
	docker build -t big-iot-gw .
	heroku container:push web --app big-iot-gw 

.PHONY: clean
clean: ## Remove all generated artefacts
	rm -rf ./build
	docker-compose -f docker-compose.yml down -v
	docker-compose down -v

# 'help' parses the Makefile and displays the help text
help:
	@echo "Please use 'make <target>' where <target> is one of"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help