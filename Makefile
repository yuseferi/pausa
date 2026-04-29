.PHONY: dev build frontend test clean release

VERSION ?=

dev:
	wails dev

build:
	wails build

frontend:
	cd frontend && npm run build

test:
	go vet ./...
	go test -race ./...

clean:
	go clean -cache
	rm -rf dist

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "usage: make release VERSION=1.0.2"; \
		exit 1; \
	fi
	./scripts/release.sh $(VERSION)
