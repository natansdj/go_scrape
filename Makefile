ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BIN_DIR = $(ROOT_DIR)/bin

export GO111MODULE=on

build:
	@(echo "-> Compiling binary")
	@(mkdir -p $(BIN_DIR))
	go build -mod=vendor -ldflags -a -o $(BIN_DIR)/go_scrape
	@(echo "-> binary created")



