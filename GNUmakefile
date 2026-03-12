default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

testunit:
	go test -v -cover -timeout=120s -parallel=10 ./...

test: testunit

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./internal/provider/...

testintegration-read:
	APOLLO_PLATFORM_INTEGRATION=1 go test -v -run '^TestClientIntegrationReadOnly$$' ./internal/provider/...

testintegration-pql:
	APOLLO_PLATFORM_INTEGRATION=1 APOLLO_INTEGRATION_ALLOW_MUTATIONS=1 go test -v -run '^TestClientIntegrationPersistedQueryListSmoke$$' ./internal/provider/...

generate-check:
	cd tools; go generate ./...
	git diff --compact-summary --exit-code

release-snapshot:
	goreleaser release --snapshot --clean --skip=publish --skip=sign

.PHONY: fmt lint testunit test testacc testintegration-read testintegration-pql build install generate generate-check release-snapshot
