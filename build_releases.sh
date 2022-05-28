#!/bin/sh

rm -rf ./build
rm -rf ./releases

platforms=("darwin" "darwin" "linux" "linux" "linux" "windows" "windows")
archs=("arm64" "amd64" "386" "arm64" "amd64" "386" "amd64")

mkdir -p releases
tbVersion=`grep ToyboxVersion version.go | head -n 1 | cut -d '"' -f2`

for ((i=0;i<${#platforms[@]};i++))
do
	mkdir -p build/${platforms[$i]}-${archs[$i]};
	GO111MODULE=off GOOS=${platforms[$i]} GOARCH=${archs[$i]} go build -o build/${platforms[$i]}-${archs[$i]}/toybox
	cp README.md build/${platforms[$i]}-${archs[$i]}/
	zip -r releases/toybox-${tbVersion}-${platforms[$i]}-${archs[$i]}.zip build/${platforms[$i]}-${archs[$i]}/*
done