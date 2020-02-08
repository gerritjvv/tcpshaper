#!/usr/bin/env bash

CMD="$1"
shift

build () {
  go build ./...
}

report () {
  go test -coverprofile=coverage.out -v ./... && \
  go tool cover -html=coverage.out
}

test () {
  go test -v ./...
}

case "$CMD" in
  build )
    build
    ;;
  report )
    report
    ;;
  test )
    test
    ;;
  * )
    echo "report|test|build"
    exit -1
    ;;
esac