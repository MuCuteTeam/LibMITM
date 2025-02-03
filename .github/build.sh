#!/bin/bash

source .github/env.sh

go get -v -d
gomobile init
gomobile bind -v -androidapi 28 -ldflags='-s -w' ./
