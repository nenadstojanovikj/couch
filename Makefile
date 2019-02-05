.PHONY: build-arm
build-arm:
	env CC=arm-linux-gnueabihf-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags '-w' -o couch-armv7

.PHONY: build
build:
	go build -v -ldflags '-w' -o couch
