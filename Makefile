BINARY_NAME=clicq-ccee-mock-server
.DEFAULT_GOAL := run

target:
	mkdir -p ../bin/
	rm -rf ../bin/*

build: clean
	go get -u ./...
	GOARCH=amd64 GOOS=linux go build -o ../bin/${BINARY_NAME} service.go

run: build
	../bin/${BINARY_NAME}

clean:
	go clean
	rm -rf ../bin/*