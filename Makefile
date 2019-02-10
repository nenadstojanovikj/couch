.PHONY: build-arm
build-arm:
	CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -v -a -tags disable_libutp -ldflags '-w -linkmode external -extldflags "-static"' -o couch-armv7

.PHONY: build
build:
	go build -v -ldflags '-w' -o couch
