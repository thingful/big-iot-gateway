SOURCE_VERSION = $(shell git describe --tags --always --dirty --abbrev=6)
EXECUTABLE = big-iot-gateway
BUILD_DIR = build
GOBUILD = go build
BUILDFLAGS = -v -ldflags "-linkmode external -extldflags -static"
PACKAGE = .

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  build		build the executable"
	@echo "  install	run glide up to install vendor dependencies"
	@echo "  release	create target container with the release version of the app"
	@echo "  run		run binary on docker container"
	@echo "  clean		remove all generated artefacts"

default: help

.PHONY: install
install:
	dep ensure -v

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILDFLAGS) -o $(BUILD_DIR)/$(EXECUTABLE) $(PACKAGE)

.PHONY: run
run:
	docker-compose build --force-rm
	docker-compose run --service-ports big_iot_gw

.PHONY: release
release:
	docker build --force-rm -t thingful/pomelo:$(SOURCE_VERSION) -t thingful/big-iot-gateway:$(SOURCE_VERSION) .

.PHONY: clean
clean:
	rm -rf ./build
	docker-compose -f docker-compose.yml down -v
	docker-compose down -v