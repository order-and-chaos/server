BIN := order-and-chaos
SOURCE := $(wildcard *.go)

.PHONY: all remake clean

$(BIN): $(SOURCE)
	go build -o $@

all: $(BIN)

clean:
	rm -f $(BIN)

remake: clean all
