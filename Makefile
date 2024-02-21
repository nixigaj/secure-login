.PHONY: build build-release clean

default: build-release

build:
	@mkdir -p build
	@go build -o build/secure-login

build-release:
	@mkdir -p build
	@go build --ldflags "-s -w" -o build/secure-login

clean:
	@rm -rf build
