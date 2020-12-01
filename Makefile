.DEFAULT_GOAL := help

# @usage Examples
# > make build
# > make build distMode=dir buildWithModule=no
# > make build distMode=dir goos=darwin buildWithModule=no

# Parameters
#====

# Go
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

# Project
VERSION=1.0-SNAPSHOT
BINARY_NAME=service
BINARY_UNIX=$(BINARY_NAME)

# Build
BUILD_WORKSPACE=tmp
BUILD_DEST=dist
COMMIT_DATE=$(shell git log -1 --pretty=tformat:'%aI')
COMMIT_AUTHOR=$(shell git log -1 --pretty=tformat:'%ae')
COMMIT_BRANCH=$(shell git log -1 --pretty=tformat:'%T')
COMMIT_HASH=$(shell git log -1 --pretty=tformat:'%H')
BUILD_DATE=$(shell date +%FT%T%z)
LDFLAGS=-ldflags '-w -extldflags "-static"'

# Enable go module
export GO111MODULE=on

# Get input parameters
#====

# Get build architecture target
ifeq ($(goos),)
	BUILD_GOOS=linux
else
	BUILD_GOOS=$(goos)
endif

ifeq ($(goarch),)
	BUILD_GOARCH=amd64
else
	BUILD_GOARCH=$(goarch)
endif

# Get module compilation mode
ifeq ($(buildWithModule),no)
   export GO111MODULE=off
else
   export GO111MODULE=on
endif

#Get dist mode
ifeq ($(distMode), dir)
    DIST_MODE=dir
    DIR_NAME=$(BINARY_NAME)
else
	DIST_MODE=tar
	DIR_NAME=$(BINARY_NAME)-$(VERSION)
endif

#Get debug mode
ifeq ($(mode),debug)
   GO_FLAGS=-gcflags "all=-N -l"
else
   GO_FLAGS=-a -tags netgo
endif

# Get ignore missing files modes
ifeq ($(ignoreMissing),yes)
   IGNORE_MISSING=true
else
   IGNORE_MISSING=false
endif

# Get tar command
ifeq ($(shell uname -s), Darwin)
	TAR=gnutar
else
	TAR=tar
endif

# Get engine script
ifeq ($(enginePath),)
	SCRIPT_DIR=../../death-angel/scripts/service
else
	SCRIPT_DIR=$(enginePath)
endif
SCRIPT_NAME=engine.sh

# Functions
#====

define buildExecutable
#==  Build ${1} with cross-compilation for ${BUILD_GOOS}:${BUILD_GOARCH}
CGO_ENABLED=0 GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} $(GOBUILD) $(LDFLAGS) $(GO_FLAGS) -o ${1} -v
endef

# Targets
#====

help:  ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


clean: ## Clean latest build data
	#== Cleaning build
	$(GOCLEAN)
	#== Removing build files
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_WORKSPACE)
	rm -rf dist

build: clean ## Build standalone package
	$(call buildExecutable,$(BINARY_NAME))

	#== Create $(BUILD_WORKSPACE) directory as build workspace
	mkdir -p $(BUILD_WORKSPACE)/$(DIR_NAME)/

	#== Include collector executable
	cp $(BINARY_UNIX) $(BUILD_WORKSPACE)/$(DIR_NAME)

	#== Include engine script
#ifneq (,$(wildcard $(SCRIPT_DIR)/$(SCRIPT_NAME)))
#	cp $(SCRIPT_DIR)/$(SCRIPT_NAME) $(BUILD_WORKSPACE)/$(DIR_NAME)
#	chmod a+x $(BUILD_WORKSPACE)/$(DIR_NAME)/$(SCRIPT_NAME)
#else

#ifeq ($(IGNORE_MISSING), true)
#	@echo "Warning! Engine script not found in path $(SCRIPT_DIR)/$(SCRIPT_NAME)"
#else
#	$(error No engine script in path $(SCRIPT_DIR)/$(SCRIPT_NAME))
#endif
#	@echo
#endif

	#== Create target data
	mkdir -p $(BUILD_DEST)
ifeq ($(DIST_MODE), tar)
	$(TAR) -cpf - -C $(BUILD_WORKSPACE) $(DIR_NAME) | xz -9 - > dist/$(DIR_NAME).tar.xz
else ifeq ($(DIST_MODE), dir)
	cp -r $(BUILD_WORKSPACE)/$(BINARY_UNIX)* $(BUILD_DEST)
else
	$(error Distribution mode $(DIST_MODE) is not supported)
endif
	#== Build executed correctly with dist mode [$(DIST_MODE)]
