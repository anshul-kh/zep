# Paths
BIN_DIR = bin
CLI_BINARY = $(BIN_DIR)/zep
Test_BINARY = $(BIN_DIR)/test
DAEMON_BINARY = $(BIN_DIR)/zep-daemon
INSTALL_BIN_DIR = /usr/local/bin
INSTALLED_DAEMON = $(INSTALL_BIN_DIR)/zep-daemon
INSTALLED_CLI = $(INSTALL_BIN_DIR)/zep
SERVICE_FILE = scripts/zep-daemon.service
SERVICE_DEST = /etc/systemd/system/zep-daemon.service


# Build Test 
build-test:
	@echo "Building Test..."
	mkdir -p $(BIN_DIR)
	go build -o $(Test_BINARY) ./cmd/test

# Build CLI
build-cli:
	@echo "Building CLI..."
	mkdir -p $(BIN_DIR)
	go build -o $(CLI_BINARY) ./cmd/cli
	sudo cp $(CLI_BINARY) $(INSTALLED_CLI)
	sudo chmod +x $(INSTALLED_CLI)

# Build Daemon
build-daemon:
	@echo "Building Daemon..."
	mkdir -p $(BIN_DIR)
	go build -o $(DAEMON_BINARY) ./cmd/daemon
	

# Build both
build: build-cli build-daemon build-test

# Install daemon binary into /usr/local/bin
install-daemon-bin: build-daemon
	@echo "Installing zep-daemon binary to $(INSTALL_BIN_DIR)..."
	sudo cp $(DAEMON_BINARY) $(INSTALLED_DAEMON)
	sudo chmod +x $(INSTALLED_DAEMON)

# Install systemd service
install-service: install-daemon-bin
	@echo "Installing systemd service..."
	sudo cp $(SERVICE_FILE) $(SERVICE_DEST)
	sudo chmod 644 $(SERVICE_DEST)
	sudo systemctl daemon-reload
	sudo systemctl enable zep-daemon
	sudo systemctl start zep-daemon
	@echo "Service zep-daemon started."

# Uninstall systemd service
uninstall-service:
	@echo "Stopping and removing systemd service..."
	- sudo systemctl stop zep-daemon
	- sudo systemctl disable zep-daemon
	- sudo rm -f $(SERVICE_DEST)
	- sudo rm -f $(INSTALLED_DAEMON)
	sudo systemctl daemon-reload
	@echo "Service zep-daemon removed."

# Restart the service
restart-service:
	sudo systemctl restart zep-daemon

# Check status
status-service:
	systemctl status zep-daemon

# Clean built binaries
clean:
	rm -rf $(BIN_DIR)

.PHONY: build build-cli build-daemon install-daemon-bin install-service uninstall-service restart-service status-service clean
