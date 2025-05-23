.PHONY: build-js-ui
build-js-ui:
	cd web/ui/js-app && yarn install && yarn build

.PHONY: build-js-ui-live-reload
build-js-ui-live-reload:
	cd web/ui/js-app && yarn install && yarn start

.PHONY: test-go
test-go:
	go test ./... -tags=dev

.PHONY: test-go-excl-integration
test-go-excl-integration:
	go test ./... -tags=dev -short

.PHONY: lint-go
lint-go:
	golangci-lint run ./...

.PHONY: security-go
security-go:
	govulncheck ./...

.PHONY: build-go
build-go:
	go build ./cmd/mdb-tool

.PHONY: build-go-dev
build-go-dev:
	go build -tags=dev ./cmd/mdb-tool
