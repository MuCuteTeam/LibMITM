#!/bin/bash

source .github/env.sh

go get -v -d
go install -v github.com/sagernet/gomobile/cmd/gomobile@v0.0.0-20221130124640-349ebaa752ca
gomobile init
go install -v github.com/sagernet/gomobile/cmd/gobind@v0.0.0-20221130124640-349ebaa752ca

BUILD="../libcore_build"

rm -rf $BUILD/android \
  $BUILD/java \
  $BUILD/javac-output \
  $BUILD/src*

gomobile bind -v -cache $(realpath $BUILD) -androidapi 28 -trimpath -tags='disable_debug' -ldflags='-s -w -buildid=' . || exit 1
rm -r libcore-sources.jar

proj=../SagerNet/app/libs
if [ -d $proj ]; then
  cp -f libcore.aar $proj
  echo ">> install $(realpath $proj)/libcore.aar"
fi
