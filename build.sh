#!/usr/bin/env bash

CMD="$1"
shift

build() {
  echo ">>>>>>>>>> building <<<<<<<<<<"
  go build -i ./... &&
    echo ">>>>>>>>>> linting <<<<<<<<<<"
  $(go list -f {{.Target}} golang.org/x/lint/golint) ./... &&
    echo ">>>>>>>>>> vetting <<<<<<<<<<"
  go vet ./...
  echo ">>>>>>>>>> end <<<<<<<<<<"
}

report() {
  go test -coverprofile=coverage.out -v ./... &&
    go tool cover -html=coverage.out
}

test() {
  go test -coverprofile=coverage.out -v ./...
}

netxtest () {

  go run main.go -count 100 && \
  go run main.go -count 100 -limit 0 && \
  go run main.go -count 100 -limit-conn 0

}
case "$CMD" in
build)
  build
  ;;
report)
  report
  ;;
test)
  test
  ;;
netxtest )
  netxtest
  ;;
*)
  echo "report|test|build"
  exit -1
  ;;
esac
