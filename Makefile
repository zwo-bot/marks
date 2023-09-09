SOURCES			= $(shell find . -name '*.go')
BINARY			= "bookmarks"


default: build.local

.PHONY: clean
clean:
	rm -rf build

.PHONY: test
test:
	go test -v -race -cover $(GOPKGS)

.PHONY: fmt
fmt: $(SOURCES)
	gofmt -l -w -s $(SOURCES)

.PHONY: lint
lint:
	golangci-lint run -v --enable gofmt

.PHONY: build.local
build.local: build/$(BINARY)

.PHONY: build.linux
build.linux: build/linux/amd64/$(BINARY)

.PHONY: build.linux.amd64
build.linux.amd64: build/linux/amd64/$(BINARY)

.PHONY: build.linux.arm64
build.linux.arm64: build/linux/arm64/$(BINARY)

build/$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build/linux/amd64/$(BINARY): $(SOURCES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/linux/amd64/$(BINARY) -ldflags "$(LDFLAGS)" .

build/linux/arm64/$(BINARY): $(SOURCES)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/linux/arm64/$(BINARY) -ldflags "$(LDFLAGS)" .