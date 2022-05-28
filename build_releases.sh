#!/bin/sh

rm -rf ./build
rm -rf ./releases

echo "Building README HTML file..."
mkdir build
cmark-gfm README.md > build/README.html

platforms=("darwin" "darwin" "linux" "linux" "linux" "windows" "windows")
archs=("arm64" "amd64" "386" "arm64" "amd64" "386" "amd64")

mkdir -p releases
tbVersion=`grep ToyboxVersion version.go | head -n 1 | cut -d '"' -f2`

echo "Cross-building version ${tbVersion}..."
for ((i=0;i<${#platforms[@]};i++))
do
	mkdir -p build/${platforms[$i]}-${archs[$i]};
	GO111MODULE=off GOOS=${platforms[$i]} GOARCH=${archs[$i]} go build -o build/${platforms[$i]}-${archs[$i]}/toybox
	cp build/README.html build/${platforms[$i]}-${archs[$i]}/

	if [ ${platforms[$i]} = 'windows' ]
	then
		mv build/${platforms[$i]}-${archs[$i]}/toybox build/${platforms[$i]}-${archs[$i]}/toybox.exe
	fi

	cd build/${platforms[$i]}-${archs[$i]}/
	zip -r ../../releases/toybox-${tbVersion}-${platforms[$i]}-${archs[$i]}.zip ./*
	cd ../..
done

echo "Building .msi..."
wixl -v deploy/toybox.wxs -o releases/toybox.msi