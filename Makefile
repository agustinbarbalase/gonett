BIN_DIR := bin
BINARY := gonett

.PHONY: all build clean install

all: build

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY) ./cmd/gonett
	@echo "✓ Built $(BIN_DIR)/$(BINARY)"

install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BIN_DIR)/$(BINARY) /usr/local/bin/
	@echo "✓ Installed"

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@echo "✓ Cleaned"

test:
	@go test ./...
