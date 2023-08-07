APP_NAME=gfa

.PHONY: all build swag
all:
	build
build:
	mkdir -p bin && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -o bin/${APP_NAME}