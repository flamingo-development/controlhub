#!/bin/sh

go build -o ./plugins/generic.so -buildmode=plugin ./default_plugins/generic/generic.go
