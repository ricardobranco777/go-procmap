BIN	= procmap
GO	= go

.PHONY: all build test clean run

all: build

build:
	$(GO) build -o $(BIN) ./cmd/procmap

gen:
	$(RM) go.mod go.sum
	$(GO) mod init github.com/ricardobranco777/procmap
	$(GO) mod tidy

test:
	$(GO) test -v
	$(GO) vet
	staticcheck
	gofmt -s -l .
	golangci-lint run -D errcheck

clean:
	$(GO) clean -a
	$(RM) $(BIN)
