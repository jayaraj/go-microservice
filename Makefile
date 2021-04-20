BUF_VERSION := 0.41.0
BUF_INSTALL_FROM_SOURCE := false
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)
CACHE_BASE := $(HOME)/.cache/$(PROJECT)
CACHE := $(CACHE_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
CACHE_BIN := $(CACHE)/bin
CACHE_VERSIONS := $(CACHE)/versions
export PATH := $(abspath $(CACHE_BIN)):$(PATH)
export GOBIN := $(abspath $(CACHE_BIN))
export GO111MODULE := on
BUF := $(CACHE_VERSIONS)/buf/$(BUF_VERSION)
$(BUF):
	@rm -f $(CACHE_BIN)/buf
	@mkdir -p $(CACHE_BIN)
	curl -sSL \
		"https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(UNAME_OS)-$(UNAME_ARCH)" \
		-o "$(CACHE_BIN)/buf"
	chmod +x "$(CACHE_BIN)/buf"
	@rm -rf $(dir $(BUF))
	@mkdir -p $(dir $(BUF))
	@touch $(BUF)

.PHONY: ci
ci: $(BUF)

.PHONY: generate
generate:
	buf beta mod update
	buf lint
	buf generate
	mv temporary/proto/* proto/openapi
	rm -r temporary
	statik -m -f -src ./proto/openapi/

.PHONY: setup
setup:
	go mod tidy
	go get \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
        google.golang.org/protobuf/cmd/protoc-gen-go \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc \
        github.com/rakyll/statik
