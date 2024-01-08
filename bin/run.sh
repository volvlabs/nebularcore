#!/bin/bash

# Build and run the Go server
go build -o ./bin .

# Watch for file changes and restart the server
reflex -r '\.go$' -s -- sh -c 'rm gitlab.com/volvlabs/nebularcore; go build .; kill -f gitlab.com/volvlabs/nebularcore; ./gitlab.com/volvlabs/nebularcore serve'
