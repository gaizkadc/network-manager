#
# Copyright (C) 2018 Nalej Group - All Rights Reserved
#
# Makefile for Nalej projects. It provides build, test, and package targets.
#

# Name of the target applications to be built
APPS=grpcserver

# Target directory to store binaries and results
TARGET=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Docker configuration
DOCKER_REPO=nalej
VERSION=$(shell cat version)



# Use ldflags to pass commit and branch information
# TODO: Integrate this into the compilation process
# LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"
# Build information
#COMMIT=$(shell git rev-parse HEAD)
#BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

COVERAGE_FILE=$(TARGET)/coverage.out

.PHONY: all
all: dep build test image

.PHONY: dep
dep:
	$(info >>> Updating dependencies...)
	dep ensure -v

test-all: test test-race test-coverage

.PHONY: test test-race test-coverage
test:
	$(info >>> Launching tests...)
	$(GOTEST) ./...

test-race:
	$(info >>> Launching tests... (Race detector enabled))
	$(GOTEST) -race ./...

test-coverage:
	$(info >>> Launching tests... (Coverage enabled))
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic  ./...

# Check the codestyle using gometalinter
.PHONY: checkstyle
checkstyle:
	gometalinter --disable-all --enable=golint --enable=vet --enable=errcheck --enable=goconst --vendor ./...

.PHONY: clean
clean:
	$(info >>> Cleaning project...)
	$(GOCLEAN)
	rm -Rf $(TARGET)

.PHONY: build-all build build-linux
build-all: build build-linux

build:
	$(info >>> Building ...)
	for app in $(APPS); do \
            $(GOBUILD) -o $(TARGET)/"$$app" ./cmd/"$$app" ; \
	done

# Cross compilation to obtain a linux binary
build-linux:
	$(info >>> Bulding for Linux...)
	for app in $(APPS); do \
    	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(TARGET)/linux_amd64/"$$app" ./cmd/"$$app" ; \
	done

# Package all images and components
.PHONY: image image-create-dir create-image
image: build-linux image-create-dir create-image

image-create-dir:
	mkdir -p $(TARGET)/images

create-image:
	$(info >>> Creating images ...)
	for app in $(APPS); do \
        echo Create image of app $$app ; \
        if [ -f components/"$$app"/Dockerfile ]; then \
            mkdir -p $(TARGET)/images/"$$app" ; \
            docker build --no-cache -t $(DOCKER_REPO)/"$$app":$(VERSION) -f components/"$$app"/Dockerfile $(TARGET)/linux_amd64 ; \
            docker save $(DOCKER_REPO)/"$$app" > $(TARGET)/images/"$$app"/image.tar ; \
            docker rmi $(DOCKER_REPO)/"$$app"; \
            cd $(TARGET)/images/"$$app"/ && tar cvzf "$$app".tar.gz * && cd - ; \
        else  \
            echo $$app has no Dockerfile ; \
        fi ; \
    done

# Publish the image
publish: image publish-image

publish-image:
	$(info >>> Publish images into Docker Hub ...)
	$(info >>> Plese type your credentials ...)
	docker login
	for app in $(APPS); do \
	    if [ -f $(TARGET)/images/"$$app"/image.tar ]; then \
	        docker push $(DOCKER_REPO)/"$$app":$(VERSION) ; \
	    else \
	        echo $$app has no image to be pushed ; \
	    fi ; \
   	    echo  Publish image of app $$app ; \
    done ; \
    docker logout ; \


