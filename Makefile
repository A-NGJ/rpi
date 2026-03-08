.PHONY: build test clean install uninstall

build:
	go build -o bin/rpi ./cmd/rpi

test:
	go test ./...

clean:
	rm -f bin/rpi

install: build
	mkdir -p $(HOME)/.local/bin
	ln -sf $(CURDIR)/bin/rpi-init $(HOME)/.local/bin/rpi-init
	ln -sf $(CURDIR)/bin/rpi $(HOME)/.local/bin/rpi
	@echo "Installed rpi-init and rpi to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH"

uninstall:
	rm -f $(HOME)/.local/bin/rpi-init
	rm -f $(HOME)/.local/bin/rpi
	@echo "Removed rpi-init and rpi from ~/.local/bin/"
