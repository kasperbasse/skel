BIN       := skel
GO        := go
GOFLAGS   :=
MEM_PKG   ?= ./cmd
MEM_RUNS  ?= 25
MEM_PKGS  ?= ./cmd ./internal/scanner ./internal/restore
MEM_REPORT_DIR ?= memreports
MEM_BASELINE_DIR ?= memreports-baseline

.PHONY: all build test lint vet fmt check memcheck memcheck-loop memcheck-report memcheck-baseline memcheck-delta ci-local tidy-check vulncheck build-darwin clean install

all: build

## build: compile the binary into ./skel
build:
	$(GO) build $(GOFLAGS) -o $(BIN) .

## install: install the binary to $GOPATH/bin
install:
	$(GO) install $(GOFLAGS) .

## test: run all tests with race detector
test:
	$(GO) test -race ./...

## test-v: run all tests verbosely
test-v:
	$(GO) test -race -v ./...

## lint: run golangci-lint (requires golangci-lint to be installed)
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## fmt: format all Go source files
fmt:
	$(GO) fmt ./...

## check: vet + lint + test (core local quality check)
check: vet lint test

## tidy-check: ensure go.mod/go.sum are tidy and committed
tidy-check:
	$(GO) mod tidy
	git diff --exit-code -- go.mod go.sum

## vulncheck: run govulncheck with pinned version
vulncheck:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

## build-darwin: build darwin arm64 and amd64 test binaries
build-darwin:
	GOOS=darwin GOARCH=arm64 $(GO) build -o skel-arm64 .
	GOOS=darwin GOARCH=amd64 $(GO) build -o skel-amd64 .

## ci-local: run the full CI-equivalent gate locally
ci-local:
	$(GO) mod verify
	$(MAKE) tidy-check
	$(MAKE) vet
	$(MAKE) lint
	$(GO) test -v -race ./...
	$(MAKE) vulncheck
	$(MAKE) build-darwin

## memcheck: run tests with heap profiling and show top allocators
memcheck:
	rm -f mem.out
	$(GO) test -count=1 -run . -memprofile=mem.out $(MEM_PKG)
	$(GO) tool pprof -top mem.out

## memcheck-loop: run repeated tests and show in-use + alloc heap profiles
memcheck-loop:
	rm -f mem.out
	$(GO) test -count=$(MEM_RUNS) -run . -memprofile=mem.out $(MEM_PKG)
	$(GO) tool pprof -sample_index=inuse_space -top mem.out
	$(GO) tool pprof -sample_index=alloc_space -top mem.out

## memcheck-report: profile multiple packages and write per-package reports
memcheck-report:
	rm -rf $(MEM_REPORT_DIR)
	mkdir -p $(MEM_REPORT_DIR)
	@for pkg in $(MEM_PKGS); do \
		name=$$(echo "$$pkg" | sed 's|^\./||; s|/|_|g'); \
		out="$(MEM_REPORT_DIR)/mem-$$name.out"; \
		txt="$(MEM_REPORT_DIR)/mem-$$name.txt"; \
		echo "Profiling $$pkg -> $$txt"; \
		$(GO) test -count=$(MEM_RUNS) -run . -memprofile="$$out" "$$pkg" >/dev/null; \
		{ \
			echo "# $$pkg"; \
			echo; \
			echo "## inuse_space"; \
			$(GO) tool pprof -sample_index=inuse_space -top "$$out"; \
			echo; \
			echo "## alloc_space"; \
			$(GO) tool pprof -sample_index=alloc_space -top "$$out"; \
		} > "$$txt"; \
	done
	@echo "Reports written to $(MEM_REPORT_DIR)/"

## memcheck-baseline: capture and store current memreports as comparison baseline
memcheck-baseline: memcheck-report
	rm -rf $(MEM_BASELINE_DIR)
	cp -R $(MEM_REPORT_DIR) $(MEM_BASELINE_DIR)
	@echo "Baseline saved to $(MEM_BASELINE_DIR)/"

## memcheck-delta: compare current memreports against baseline totals
memcheck-delta:
	@test -d "$(MEM_REPORT_DIR)" || (echo "missing $(MEM_REPORT_DIR)/ (run make memcheck-report first)" && exit 1)
	@test -d "$(MEM_BASELINE_DIR)" || (echo "missing $(MEM_BASELINE_DIR)/ (run make memcheck-baseline first)" && exit 1)
	@echo "Comparing $(MEM_REPORT_DIR)/ vs $(MEM_BASELINE_DIR)/"
	@for newf in $(MEM_REPORT_DIR)/mem-*.txt; do \
		basef="$(MEM_BASELINE_DIR)/$$(basename "$$newf")"; \
		name="$$(basename "$$newf" .txt)"; \
		if [ ! -f "$$basef" ]; then \
			echo "$$name: missing baseline report"; \
			continue; \
		fi; \
		set -- $$(awk '/^## inuse_space/{m="i";next} /^## alloc_space/{m="a";next} /^Showing nodes accounting for /{v=$$5; gsub(/[^0-9.]/,"",v); if(m=="i"&&i=="")i=v; else if(m=="a"&&a=="")a=v} END{if(i=="")i=0; if(a=="")a=0; printf "%.2f %.2f", i, a}' "$$basef"); \
		base_inuse="$$1"; base_alloc="$$2"; \
		set -- $$(awk '/^## inuse_space/{m="i";next} /^## alloc_space/{m="a";next} /^Showing nodes accounting for /{v=$$5; gsub(/[^0-9.]/,"",v); if(m=="i"&&i=="")i=v; else if(m=="a"&&a=="")a=v} END{if(i=="")i=0; if(a=="")a=0; printf "%.2f %.2f", i, a}' "$$newf"); \
		new_inuse="$$1"; new_alloc="$$2"; \
		inuse_pct=$$(awk -v b="$$base_inuse" -v n="$$new_inuse" 'BEGIN{if(b==0){print "n/a"}else{printf "%+.1f%%", ((n-b)/b)*100}}'); \
		alloc_pct=$$(awk -v b="$$base_alloc" -v n="$$new_alloc" 'BEGIN{if(b==0){print "n/a"}else{printf "%+.1f%%", ((n-b)/b)*100}}'); \
		echo "$$name  inuse: $$base_inuse -> $$new_inuse kB ($$inuse_pct)  alloc: $$base_alloc -> $$new_alloc kB ($$alloc_pct)"; \
	done

## clean: remove build artifacts
clean:
	rm -f $(BIN) $(BIN)-arm64 $(BIN)-amd64 mem.out Brewfile cmd.test coverage.out profile.test refactor.test scanner.test restore.test
	rm -rf $(MEM_BASELINE_DIR)
	rm -rf $(MEM_REPORT_DIR)

## help: print this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'

