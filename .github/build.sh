#!/bin/bash

source .github/env.sh

go get -v -d
go install -v github.com/sagernet/gomobile/cmd/gomobile@v0.0.0-20221130124640-349ebaa752ca
gomobile init
go install -v github.com/sagernet/gomobile/cmd/gobind@v0.0.0-20221130124640-349ebaa752ca

gomobile bind -v -cache $(realpath $BUILD) -androidapi 28 -trimpath -tags='disable_debug' -ldflags='-s -w -buildid=' . || exit 1
