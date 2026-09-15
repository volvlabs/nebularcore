#!/bin/bash

# Build and run the Go server
go build -o ./bin .

# Watch for file changes and restart the server
reflex -r '\.go$' -s -- sh -c 'rm github.com/volvlabs/nebularcore; go build .; kill -f github.com/volvlabs/nebularcore; ./github.com/volvlabs/nebularcore serve'
