#!/bin/bash

source .github/env.sh

go get -v -d
go install -v golang.org/x/mobile/cmd/gomobile@v0.0.0-20221110043201-43a038452099
go install -v golang.org/x/mobile/cmd/gobind@v0.0.0-20221110043201-43a038452099
gomobile init
gomobile bind -v -androidapi 28 -ldflags='-s -w' ./
