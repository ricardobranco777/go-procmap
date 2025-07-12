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
