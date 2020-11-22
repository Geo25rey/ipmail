#!/bin/bash

# This script creates packages for Linux, Windows, and Mac from a Linux Host system
# Depends on osxcross using Mac OS 11.0 Sdk
# Depends on MinGW-w64

mkdir build
cd build

go get fyne.io/fyne/cmd/fyne

# Package for Linux
fyne package -os linux -appID io.ipmail -release -sourceDir ..

# Package for Mac
CC=x86_64-apple-darwin20-clang fyne package -os darwin -appID io.ipmail -release -sourceDir ..

# Package for Windows
CC=x86_64-w64-mingw32-gcc fyne package -os windows -appID io.ipmail -release -sourceDir ..

cd ..

mv ipmail* build/