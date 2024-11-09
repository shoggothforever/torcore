BINARY_DIR=cmd/bitctl
BINARY_NAME=bitctl.exe
CORBRA=cobra-cli

all: build

build:
	go build -o $(BINARY_NAME) $(BINARY_DIR)/main.go

run: build
	./$(BINARY_NAME) start

clean:
	rm -f $(BINARY_DIR)/$(BINARY_NAME)

status:
	./$(BINARY_NAME) status
