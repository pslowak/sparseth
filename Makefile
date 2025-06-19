APP_NAME = sparseth
MAIN_PKG = ./cmd/sparseth

BUILD_DIR = ./build
CONTRACTS_DIR = ./contracts

.PHONY: all build test clean

all: build

build:
	solc --abi --storage-layout --overwrite -o $(BUILD_DIR) $(CONTRACTS_DIR)/*.sol
	go build -o $(BUILD_DIR)/bin/$(APP_NAME) $(MAIN_PKG)
test:
	go test ./... -v

clean:
	rm -rf $(BUILD_DIR)
