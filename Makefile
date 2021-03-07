.PHONY: all exes clean
.DEFAULT_GOAL: all

# Build Info
VERSION := $(shell cat "./VERSION" 2>/dev/null)
GIT_REVISION := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date +%Y/%m/%d)

# Compiler & Linker
GO_LDFLAGS += -X 'main.Version=$(VERSION)'
GO_LDFLAGS += -X 'main.BuildTime=$(BUILD_TIME)'
GO_LDFLAGS += -X 'main.GitRevision=$(GIT_REVISION)'
GO_LDFLAGS += -extldflags '-static'
GO_LDFLAGS += -s -w
GO_FLAGS += -ldflags "$(GO_LDFLAGS)"

# Binaries
MAIN_GO := $(shell find . -type f -name 'main.go' -print)
MAIN_EXES := $(foreach exe, $(patsubst ./cmd/%/main.go, %, $(MAIN_GO)), ./cmd/$(exe)/$(exe))

define dep_exe
$(1): $(dir $(1))/main.go
endef
$(foreach exe, $(MAIN_EXES), $(eval $(call dep_exe, $(exe))))

# Binaries deps
exes: $(MAIN_EXES)
$(MAIN_EXES):
	CGO_ENABLED=0 go build $(GO_FLAGS) -o $@ ./$(@D)

# Clean target
clean:
	rm -rf $(EXES)
	go clean ./...

# Install target
install: exes
	cp $(MAIN_EXES) /usr/bin/

# Default goal
all: exes

