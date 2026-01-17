.PHONY: build test run clean scaffold

BINARY_NAME=mend
BUILD_DIR=bin
DATA_DIR=test_data

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	go test ./...

scaffold:
	./scripts/scaffold_data.sh

run: scaffold
	go run . $(DATA_DIR)

clean:
	rm -rf $(BUILD_DIR)

clean-data:
	rm -rf $(DATA_DIR)
