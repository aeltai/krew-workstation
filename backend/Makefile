# Build for current platform
build:
	CGO_ENABLED=0 go build -o krew-manager .

# Build for macOS Intel (amd64)
build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o krew-manager-darwin-amd64 .

# Build for macOS Apple Silicon (arm64)
build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o krew-manager-darwin-arm64 .

# Build for macOS (auto-detect: arm64 on M1/M2, amd64 on Intel)
build-mac:
	@case "$$(uname -m)" in \
		arm64) $(MAKE) build-darwin-arm64 && mv krew-manager-darwin-arm64 krew-manager ;; \
		x86_64) $(MAKE) build-darwin-amd64 && mv krew-manager-darwin-amd64 krew-manager ;; \
		*) echo "Unknown arch"; exit 1 ;; \
	esac

# Build both macOS architectures
build-darwin-all: build-darwin-amd64 build-darwin-arm64

.PHONY: build build-darwin-amd64 build-darwin-arm64 build-mac build-darwin-all
