#
# Makefile for generating gRPC stubs in Go
#

all: lint generate

.PHONY: lint generate clean
lint:
	$(info >>> Checking style...)
	protoc --lint_out=. */*.proto

generate:
	chmod +x create_stubs.sh
	./create_stubs.sh

clean:
	rm */*.pb.go
