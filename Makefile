BINDIR		:= $(CURDIR)/bin
BINNAME		?= zincsearch

GOPATH        = $(shell go env GOPATH)
GOIMPORTS     = $(GOPATH)/bin/goimports

# go option
PKG        := ./...
TAGS       :=
TESTS      := .
TESTFLAGS  :=
LDFLAGS    := -w -s
GOFLAGS    :=
SRC        := $(shell find . -type f -name '*.go' -print)

# ------------------------------------------------------------------------------
#  build

build: clean tidy fmt vet compile

tidy:
	@echo
	@echo "=== tidying ==="
	go mod tidy

.PHONY: compile
compile: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	@echo
	@echo "=== running compile ==="
	GO111MODULE=on CGO_ENABLED=0 go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) cmd/zincsearch/main.go

# Run go fmt against code
fmt:
	@echo
	@echo "=== fmt ==="
	go fmt ./...

# Run go vet against code
vet:
	@echo
	@echo "=== vet ==="
	go vet ./...

.PHONY: test-unit
test-unit:
	@echo
	@echo "=== running unit tests ==="
	GO111MODULE=on go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: clean
clean:
	@echo
	@echo "=== cleaning ==="
	rm -rf $(BINDIR)
	rm -rf data
	find ./pkg -name data -type dir|xargs rm -fR
	find ./test -name data -type dir|xargs rm -fR