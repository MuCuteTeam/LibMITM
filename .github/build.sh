#!/bin/bash

source .github/env.sh

go get -v -d
gomobile init

BUILD="../libmitm"

gomobile bind -v -androidapi 28 -ldflags='-s -w' ./
