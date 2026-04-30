APP_NAME := local-skill-manager
MAIN_PKG := ./cmd/local-skill-manager
DIST_DIR := dist
GO_CACHE_DIR := $(CURDIR)/.cache/go-build
GO_MOD_CACHE_DIR := $(CURDIR)/.cache/go-mod
GO_ENV := GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE_DIR)

.PHONY: build build-cross clean test

build:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go build -buildvcs=false -o $(APP_NAME) $(MAIN_PKG)

build-cross:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) GOOS=linux GOARCH=amd64 go build -buildvcs=false -o $(DIST_DIR)/$(APP_NAME)_linux_amd64 $(MAIN_PKG)
	$(GO_ENV) GOOS=linux GOARCH=arm64 go build -buildvcs=false -o $(DIST_DIR)/$(APP_NAME)_linux_arm64 $(MAIN_PKG)
	$(GO_ENV) GOOS=darwin GOARCH=amd64 go build -buildvcs=false -o $(DIST_DIR)/$(APP_NAME)_darwin_amd64 $(MAIN_PKG)
	$(GO_ENV) GOOS=darwin GOARCH=arm64 go build -buildvcs=false -o $(DIST_DIR)/$(APP_NAME)_darwin_arm64 $(MAIN_PKG)
	$(GO_ENV) GOOS=windows GOARCH=amd64 go build -buildvcs=false -o $(DIST_DIR)/$(APP_NAME)_windows_amd64.exe $(MAIN_PKG)
	$(GO_ENV) GOOS=windows GOARCH=arm64 go build -buildvcs=false -o $(DIST_DIR)/$(APP_NAME)_windows_arm64.exe $(MAIN_PKG)
	cd $(DIST_DIR) && zip -q $(APP_NAME)_linux_amd64.zip $(APP_NAME)_linux_amd64
	cd $(DIST_DIR) && zip -q $(APP_NAME)_linux_arm64.zip $(APP_NAME)_linux_arm64
	cd $(DIST_DIR) && zip -q $(APP_NAME)_darwin_amd64.zip $(APP_NAME)_darwin_amd64
	cd $(DIST_DIR) && zip -q $(APP_NAME)_darwin_arm64.zip $(APP_NAME)_darwin_arm64
	cd $(DIST_DIR) && zip -q $(APP_NAME)_windows_amd64.zip $(APP_NAME)_windows_amd64.exe
	cd $(DIST_DIR) && zip -q $(APP_NAME)_windows_arm64.zip $(APP_NAME)_windows_arm64.exe

test:
	mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	$(GO_ENV) go test ./...

clean:
	rm -rf $(DIST_DIR) $(APP_NAME)
