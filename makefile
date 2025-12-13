.PHONY: help test test-unit test-acc

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  test       Run all tests (unit + acceptance)"
	@echo "  test-unit  Run unit tests"
	@echo "  test-acc   Run acceptance tests (requires UPTIME_API_KEY)"

test: test-unit test-acc

test-unit:
	go test ./...

test-acc:
	go test -tags=acc -run 'TestAcc' -v ./...
