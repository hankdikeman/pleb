ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BINDIR := $(ROOTDIR)/prod/bin

PROTOC := protoc --go_out=./ --go_opt=paths=source_relative
GRPC-PROTOC := $(PROTOC) --go-grpc_out=./ --go-grpc_opt=paths=source_relative

###
# Direct build targets to build standalone binaries
###

horrea-proto:
	$(GRPC-PROTOC) prod/horrea/pb/horrea.proto
horrea: horrea-proto
	go build -C prod/horrea/main -o $(BINDIR)/horrea

pleb-proto:
	$(GRPC-PROTOC) prod/pleb/pb/pleb.proto
pleb: pleb-proto
	go build -C prod/pleb/main -o $(BINDIR)/pleb

fora-proto:
	$(GRPC-PROTOC) prod/fora/pb/fora.proto
fora: fora-proto
	go build -C prod/fora/main -o $(BINDIR)/fora

fabricae-proto:
	$(GRPC-PROTOC) prod/fabricae/pb/fabricae.proto
fabricae: fabricae-proto
	go build -C prod/fabricae/main -o $(BINDIR)/fabricae

caesar-proto:
	$(GRPC-PROTOC) prod/caesar/pb/caesar.proto
caesar: caesar-proto
	go build -C prod/caesar/main -o $(BINDIR)/caesar

senator-proto:
	$(GRPC-PROTOC) prod/senator/pb/senator.proto
senator: senator-proto
	go build -C prod/senator/main -o $(BINDIR)/senator

all-proto: horrea-proto # pleb-proto fora-proto fabricae-proto caesar-proto senator-proto TODO uncomment when all protos done

all: horrea pleb fora fabricae caesar senator

###
# Local test targets
###

.PHONY: horrea-test
horrea-test: horrea
	go test -C prod/horrea/main ./...

###
# Dockerfile targets
###

# TODO not sure if I need these yet but would be like: `sudo docker build -t [x]-server prod/[x]`
# I don't know what testing I would need the actual image for. Maybe grander component tests?

# builds and launches all docker containers for
.PHONY: serverup
serverup: all-proto
	sudo docker-compose up -d

###
# Utility targets
###

.PHONY: format
format:
	gofmt -w -s .

# remove all generated proto files and binaries
.PHONY: clean
clean:
	find . -name *.pb.go -delete
	rm -f $(BINDIR)/*

# removes all generated docker images too
.PHONY: deep-clean
deep-clean: clean
	sudo docker system prune -af
