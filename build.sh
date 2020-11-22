#!/bin/bash

mkdir build
cd build
rm -r .
go build ..
mv main ipmail
cd ..

