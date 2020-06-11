ARCHS ?= amd64
GOOS ?= windows

BINARY=kuassh.exe
VERSION=`git describe --tags`
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"


build:
	GOOS=$(GOOS) GOARCH=$(ARCHS) go build ${LDFLAGS} -o ${BINARY} _example/conn.go

install:
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi