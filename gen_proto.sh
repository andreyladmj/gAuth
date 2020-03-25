#!/bin/bash
protoc pkg/grpc/userpb/userpb.proto --go_out=plugins=grpc:.
