.PHONY: build clean dev run restart deploy test test-race vet lint fmt fmt-check ci stop-port

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
# Go sources excluding vendored frontend deps and the ignored scratch dir.
GO_SRC := $(shell find . -name '*.go' -not -path './scratch/*' -not -path './frontend/*')

# build compiles both binaries atomically: each goes to a .tmp first and is only
# promoted (mv) after BOTH compile, so an interrupted or half-failed build can
# never leave a corrupt or mixed-version pair in the deploy dir.
build:
	go build -ldflags="-X main.BuildCommit=$(GIT_COMMIT)" -o swarmops.tmp .
	go build -o quota-proxy.tmp ./cmd/quota-proxy
	mv -f swarmops.tmp swarmops
	mv -f quota-proxy.tmp quota-proxy

dev:
	go run .

clean:
	rm -f swarmops swarmops.tmp swarmops.prev
	rm -f quota-proxy quota-proxy.tmp quota-proxy.prev
	rm -f .deploy.lock
	rm -f swarmops.db
	rm -f swarmops-test*.db*

# stop-port gracefully stops whatever is listening on 8080
stop-port:
	@pid=$$(fuser 8080/tcp 2>/dev/null | xargs); \
	if [ -n "$$pid" ]; then \
		echo "Port 8080 in use by PID $$pid — sending SIGTERM..."; \
		kill $$pid 2>/dev/null || true; \
		for i in 1 2 3 4 5 6; do \
			fuser 8080/tcp >/dev/null 2>&1 || { echo "Port 8080 free."; exit 0; }; \
			sleep 1; \
		done; \
		echo "Still running after 6s — sending SIGKILL..."; \
		kill -9 $$pid 2>/dev/null || true; \
		sleep 1; \
	fi

run: build stop-port
	./swarmops

# restart builds + restarts the local service. NOTE: this does NOT sync to
# origin/main — it builds whatever is checked out. For a real production deploy
# use `make deploy` (or `swarmops redeploy`), which goes through scripts/deploy.sh.
restart: build stop-port
	systemctl --user restart swarmops
	@for i in 1 2 3 4 5; do \
		sleep 1; \
		if curl -sf http://localhost:8080/api/dashboard/stats >/dev/null 2>&1; then \
			echo "SwarmOps restarted successfully ($(GIT_COMMIT))"; \
			exit 0; \
		fi; \
	done; \
	echo "WARNING: SwarmOps may not have started — check: journalctl --user -u swarmops -n 20"

# deploy is the ONE sanctioned production deploy path: scripts/deploy.sh resets
# ~/swarmops to origin/main, rebuilds, restarts, health-checks, and rolls back
# on failure. Pass ARGS=--force to deploy over a dirty deploy tree.
deploy:
	bash scripts/deploy.sh $(ARGS)

test:
	go test -timeout 120s -count=1 ./...

test-race:
	go test -race -timeout 120s -count=1 ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run ./... \
		|| echo "golangci-lint not installed — skipping"

fmt:
	@gofmt -w $(GO_SRC)

# fmt-check fails (non-zero) if any tracked Go source is not gofmt-clean.
fmt-check:
	@unformatted=$$(gofmt -l $(GO_SRC)); \
	if [ -n "$$unformatted" ]; then \
		echo "Not gofmt-clean (run 'make fmt'):"; echo "$$unformatted"; exit 1; \
	fi

# ci is the merge gate: formatting + vet + race tests must all pass.
ci: fmt-check vet test-race
