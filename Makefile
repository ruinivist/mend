.PHONY: build test run clean

BINARY_NAME=mend
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	go test ./...

run:
	go run .

clean:
	rm -rf $(BUILD_DIR)
