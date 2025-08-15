BIN	= procmap
GO	= go

.PHONY: all build test clean run

all: build

build:
	$(GO) build -o $(BIN) ./cmd/procmap

test:
	$(GO) test -v
	$(GO) vet
	staticcheck
	gofmt -s -l .

clean:
	$(GO) clean -a
	$(RM) $(BIN)

.PHONY: gen
gen:
	$(RM) go.mod go.sum
	$(GO) mod init github.com/ricardobranco777/go-$(BIN)
	$(GO) mod tidy
