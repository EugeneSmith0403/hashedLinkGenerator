#!/bin/bash

# Run tests (tests use .env.test automatically)
go test -v ./cmd/... "$@"
