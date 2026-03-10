.PHONY: build test clean install uninstall

build:
	go build -o bin/rpi ./cmd/rpi

test:
	go test ./...

clean:
	rm -f bin/rpi

install: build
	mkdir -p $(HOME)/.local/bin
	rm -f $(HOME)/.local/bin/rpi
	cp $(CURDIR)/bin/rpi $(HOME)/.local/bin/rpi
	@echo "Installed rpi to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH"

uninstall:
	rm -f $(HOME)/.local/bin/rpi
	@echo "Removed rpi from ~/.local/bin/"
