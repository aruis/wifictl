BINARY := wifictl
DIST_DIR := dist

.PHONY: build test clean

build:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY) ./

test:
	go test ./...

clean:
	rm -rf $(DIST_DIR)
