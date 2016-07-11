#!/bin/sh


mkdir -p dist

GOOS=darwin  GOARCH=amd64 go build  -ldflags="-s -w" -o dist/natip.osx bin/natip/main.go  
GOOS=linux   GOARCH=amd64 go build  -ldflags="-s -w" -o dist/natip bin/natip/main.go  
GOOS=windows GOARCH=amd64 go build  -ldflags="-s -w" -o dist/natip.exe bin/natip/main.go 
