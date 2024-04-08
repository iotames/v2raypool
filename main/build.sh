#!/bin/bash

APP_VERSION=$1
GO_VERSION=$(go version)
WinFileName="v2raypool.exe"
LinuxFineName="v2raypool"

if [ "$APP_VERSION" = "" ];then
    APP_VERSION="v1.2.2"
fi

CGO_ENABLED=0
GOARCH=amd64

BuildArgs="-trimpath -ldflags \"-w -s -buildid= \
-X 'main.AppVersion=${APP_VERSION}' \
-X 'main.GoVersion=${GO_VERSION}'\" \
-gcflags=\"all=-trimpath=${PWD}\" \
-asmflags=\"all=-trimpath=${PWD}\""

sh -c "GOOS=linux go build -o $LinuxFineName ${BuildArgs} ."

sh -c "GOOS=windows go build -o $WinFileName ${BuildArgs} ."