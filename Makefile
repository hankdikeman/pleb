BINDIR := ./bin

PROTOC := protoc --go_out=./ --go_opt=paths=source_relative
GRPC-PROTOC := $(PROTOC) --go-grpc_out=./ --go-grpc_opt=paths=source_relative

horrea:
	$(GRPC-PROTOC) prod/horrea/pb/horrea.proto
	go build -C prod/horrea/main -o $(BINDIR)/horrea

pleb:
	$(GRPC-PROTOC) prod/pleb/pb/pleb.proto
	go build -C prod/pleb/main -o $(BINDIR)/pleb

fora:
	$(GRPC-PROTOC) prod/fora/pb/fora.proto
	go build -C prod/fora/main -o $(BINDIR)/fora

fabricae:
	$(GRPC-PROTOC) prod/fabricae/pb/fabricae.proto
	go build -C prod/fabricae/main -o $(BINDIR)/fabricae

caesar:
	$(GRPC-PROTOC) prod/caesar/pb/caesar.proto
	go build -C prod/caesar/main -o $(BINDIR)/caesar

senator:
	$(GRPC-PROTOC) prod/senator/pb/senator.proto
	go build -C prod/senator/main -o $(BINDIR)/senator

all: horrea pleb fora fabricae caesar senator

.PHONY: format
format:
	gofmt -w -s .

.PHONY: clean
clean:
	find . -name *.pb.go -delete
