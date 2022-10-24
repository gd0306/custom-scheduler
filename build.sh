#!/usr/bin/env bash
# shellcheck disable=SC2181
# shellcheck disable=SC2034
# shellcheck disable=SC2034

# SET SCHEDULER VERSION
VERSION=0.0.1
# SET SCHEDULER NAME
NAME=custom-scheduler
# GO LOCAL RUNTIME
GO=$GOROOT/bin/go
# SET LOCAL WORK DIRECTORY
WORK_DIR=$GOPATH/src/custom-scheduler

# BUILD BINARY FILE
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $GO build -o "$WORK_DIR"/$NAME "$WORK_DIR"/main.go
if [ $? != 0 ];then
  echo "ERROR: binary file build failed!"
  exit $?
fi

# BUILD DOCKER IMAGES （eg: custom-scheduler:0.0.1）
docker build --no-cache -t $NAME:$VERSION -f "$WORK_DIR"/Dockerfile .
docker tag $NAME:$VERSION gd306.cn/kubernetes/$NAME:$VERSION