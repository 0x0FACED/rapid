#!/bin/sh
export GOOS=windows
export GOARCH=amd64
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
cd cmd/rapid
go build -o rapid.exe main.go
fyne package -os windows -icon test.png
