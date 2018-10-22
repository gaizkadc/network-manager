#!/bin/bash

echo "Generating gRPC stubs"

for service_dir in $(ls -d */); do
    SERVICE=${service_dir%%/};
    echo "Generating protobuf for ${SERVICE}"
    protoc -I ${SERVICE}/ ${SERVICE}/*.proto --go_out=plugins=grpc:${SERVICE}
done
