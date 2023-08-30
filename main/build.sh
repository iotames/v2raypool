#!/bin/bash

APP_VERSION=$1
GO_VERSION=$(go version)
WinFileName="v2raypool.exe"
LinuxFineName="v2raypool"

if [ "$APP_VERSION" = "" ];then
    APP_VERSION="v1.0.1"
fi

CGO_ENABLED=0
GOARCH=amd64

BuildArgs="-ldflags \"-w -s \
-X 'main.AppVersion=${APP_VERSION}' \
-X 'main.GoVersion=${GO_VERSION}'\" \
-gcflags=\"all=-trimpath=${PWD}\" \
-asmflags=\"all=-trimpath=${PWD}\""

bash -c "GOOS=linux go build ${BuildArgs} -o $LinuxFineName ."

bash -c "GOOS=windows go build ${BuildArgs} -o $WinFileName ."