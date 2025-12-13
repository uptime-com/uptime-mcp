.PHONY: test e2e

test:
	go test ./...

e2e:
	go test -tags=e2e -v ./e2e/...
