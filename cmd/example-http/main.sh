#!/bin/bash
set -eux
go build -o main main.go
./main
