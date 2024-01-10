ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BINDIR := $(ROOTDIR)/prod/bin
MOUNTPOINT := /mnt/pleb

PROTOC := protoc --go_out=./ --go_opt=paths=source_relative
GRPC-PROTOC := $(PROTOC) --go-grpc_out=./ --go-grpc_opt=paths=source_relative

###
# Direct build targets to build standalone binaries
###

# distributed KV store for locking, concurrency, service discovery
fora-proto:
	$(GRPC-PROTOC) prod/fora/pb/fora.proto
fora: fora-proto
	go build -C prod/fora/main -o $(BINDIR)/fora

# slow-access storage frontend
horrea-proto:
	$(GRPC-PROTOC) prod/horrea/pb/horrea.proto
horrea: horrea-proto
	go build -C prod/horrea/main -o $(BINDIR)/horrea

# fast-access storage frontend
fabricae-proto:
	$(GRPC-PROTOC) prod/fabricae/pb/fabricae.proto
fabricae: fabricae-proto
	go build -C prod/fabricae/main -o $(BINDIR)/fabricae

# client authentication and server management
caesar-proto:
	$(GRPC-PROTOC) prod/caesar/pb/caesar.proto
caesar: caesar-proto
	go build -C prod/caesar/main -o $(BINDIR)/caesar

# client-facing file operation service
senator-proto:
	$(GRPC-PROTOC) prod/senator/pb/senator.proto
senator: senator-proto
	go build -C prod/senator/main -o $(BINDIR)/senator

# file access concurrency manager
iudex-proto:
	$(GRPC-PROTOC) prod/iudex/pb/iudex.proto
iudex: iudex-proto
	go build -C prod/iudex/main -o $(BINDIR)/iudex

# client service
pleb:
	go build -C prod/pleb/main -o $(BINDIR)/pleb

all-proto: horrea-proto fora-proto fabricae-proto iudex-proto caesar-proto senator-proto

all: horrea pleb fora fabricae caesar senator iudex

###
# Local test targets
###

.PHONY: horrea-test
horrea-test: horrea
	go test -C prod/horrea/main ./...

.PHONY: pleb-test
pleb-test: pleb
	go test -C prod/pleb/main ./...

###
# Dockerfile targets
###

# TODO not sure if I need these yet but would be like: `sudo docker build -t [x]-server prod/[x]`
# I don't know what testing I would need the actual image for. Maybe grander component tests?

# builds and launches all docker containers for remote filesystem
.PHONY: serverup
serverup: all-proto
	sudo docker-compose up --build

###
# Utility targets
###

.PHONY: format
format:
	gofmt -w -s .

# unmount test mountpoint (defaults to /mnt/pleb)
.PHONY: unmount
unmount:
	sudo umount -l $(MOUNTPOINT) || /bin/true

# remove all generated proto files and binaries
.PHONY: clean
clean:
	find . -name *.pb.go -delete
	rm -f $(BINDIR)/*

# removes all generated docker images too
.PHONY: deep-clean
deep-clean: clean
	sudo docker system prune -af
