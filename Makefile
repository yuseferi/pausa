.PHONY: dev build frontend icons test check seed-dist clean release install-local

VERSION ?=

# Go package patterns. We avoid `./...` because the frontend's npm
# dependencies (notably flatted) ship Go source under frontend/node_modules
# that is not part of this module.
GO_PKGS := ./internal/... .

dev:
	wails dev

build:
	wails build

frontend:
	cd frontend && npm run build

# Regenerate the app icon (.icns/.iconset/build/appicon.png) from
# build/icons/appicon.svg. Requires macOS built-ins only.
icons:
	./build/icons/build-icons.sh

# The Go //go:embed directive requires frontend/dist to exist. Seed an empty
# directory so backend checks work on a clean checkout (mirrors CI).
seed-dist:
	@mkdir -p frontend/dist && touch frontend/dist/.gitkeep

test: seed-dist
	go vet $(GO_PKGS)
	go test -race $(GO_PKGS)
	go build $(GO_PKGS)

check: test
	GOOS=linux CGO_ENABLED=0 go build $(GO_PKGS)
	cd frontend && npm run lint
	cd frontend && npm run test
	cd frontend && npm run build

clean:
	go clean -cache
	rm -rf dist

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "usage: make release VERSION=1.0.2"; \
		exit 1; \
	fi
	./scripts/release.sh $(VERSION)

install-local: build
	rm -rf /Applications/pausa.app
	cp -R build/bin/pausa.app /Applications/pausa.app
	@echo "Installed /Applications/pausa.app"
	@echo "If macOS blocks first launch, run: xattr -dr com.apple.quarantine /Applications/pausa.app"
