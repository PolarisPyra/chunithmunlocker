# Define the binary name
BINARY_NAME = chunithmunlocker
# Define the build directory
BUILD_DIR = build
# Define the source directory
SRC_DIR = ./src

# Default target (build for current platform)
all: build

# Detect operating system
ifeq ($(OS),Windows_NT)
    # Windows commands
    RM = del /Q
    MKDIR = if not exist "$(BUILD_DIR)" mkdir
    BUILD_WINDOWS = go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC_DIR)/.
    BUILD_LINUX = go env -w GOOS=linux GOARCH=amd64 && go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)/. && go env -w GOOS=windows GOARCH=amd64
else
    # Linux/Unix commands
    RM = rm -rf
    MKDIR = mkdir -p
    BUILD_WINDOWS = GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC_DIR)/.
    BUILD_LINUX = GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)/.
endif

# Build for Windows
build-windows: $(BUILD_DIR)
	$(BUILD_WINDOWS)

# Build for Linux
build-linux: $(BUILD_DIR)
	$(BUILD_LINUX)

# Build for both platforms
build: build-windows build-linux

# Run the binary for the current platform
run:
	go run $(SRC_DIR)/.

# Clean the build directory
clean:
	$(RM) $(BUILD_DIR)

# Ensure the build directory exists
$(BUILD_DIR):
	$(MKDIR) $(BUILD_DIR)

.PHONY: all build build-windows build-linux run clean