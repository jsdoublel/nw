BINARY_NAME=nw
MAIN_GO_FILE=nw.go

build:
	@echo "Building for current system..."
	go build -o $(BINARY_NAME) $(MAIN_GO_FILE)
	@echo "Build complete: $(BINARY_NAME)"

all: cross-compile

TARGETS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64 \
	darwin/amd64 \
	darwin/arm64

cross-compile: $(TARGETS)
	@echo "Cross-compilation finished."

$(TARGETS):
	$(eval GOOS := $(word 1,$(subst /, ,$@)))
	$(eval GOARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT := $(if $(findstring windows,$(GOOS)),.exe,))
	@echo "Building for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME)-$(GOOS)-$(GOARCH)$(EXT) $(MAIN_GO_FILE)

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	@echo "Clean complete."

.PHONY: all build clean cross-compile $(TARGETS)
