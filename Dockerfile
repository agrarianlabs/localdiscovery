FROM       golang:1.5
RUN        go get github.com/tools/godep
RUN        go get github.com/golang/lint/golint
RUN        go get golang.org/x/tools/cmd/goimports
ENV        CGO_ENABLED 0
RUN        go install -a std

MAINTAINER Guillaume J. Charmes <gcharmes@leaf.ag>

ENV        APP_DIR    $GOPATH/src/github.com/agrarianlabs/router/discover

WORKDIR    $APP_DIR

ADD        . $APP_DIR
RUN        godep go build -ldflags -d ./cmd/discover

CMD        ["./discover"]
