TARGET_NAME = proxyMon.exe
TARGET_FOLDER = ./bin
OS = windows
ARCH = amd64
SOURCE = ./cmd/app/main.go 
LDFLAGS = -H windowsgui 

.PHONY: all build clean

# Rebuild application
all: clean build 

# Build application
build: 
	mkdir -p $(TARGET_FOLDER)
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags "$(LDFLAGS)" -o $(TARGET_FOLDER)/$(TARGET_NAME) $(SOURCE)

# Clear binaries
clean: 
	rm -rf $(TARGET_FOLDER) 
