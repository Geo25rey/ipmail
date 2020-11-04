#!/bin/bash

mkdir build
cd build
rm -r .
go build ../main.go
mv main ipmail
cd ..

