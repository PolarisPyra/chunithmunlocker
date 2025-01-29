# Define the binary name
BINARY_NAME = artemis2tachi

# Define the build directory
BUILD_DIR = build

# Define the target platforms
WINDOWS = windows
LINUX = linux

# Define the architecture
ARCH = amd64

# Define the source directory
SRC_DIR = ./src

# Default target (build for current platform)
all: build

# Build for Windows
build-windows: $(BUILD_DIR)
	GOOS=$(WINDOWS) GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(WINDOWS)-$(ARCH).exe $(SRC_DIR)/.

# Build for Linux
build-linux: $(BUILD_DIR)
	GOOS=$(LINUX) GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(LINUX)-$(ARCH) $(SRC_DIR)/.

# Build for both platforms
build: build-windows build-linux

# Run the binary for the current platform
run:
	go run $(SRC_DIR)/.

# Clean the build directory
clean:
	rm -rf $(BUILD_DIR)

# Ensure the build directory exists
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)
