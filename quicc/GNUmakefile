BIN=quicc

default:: help

dev:: ## 10 Build a local build
	go build -o quicc

all:: quicc_illumos ## 10 Build all targets

quicc_illumos:: ## 10 Build for Illumos
	env GOOS=solaris GOARCH=amd64 go build -o $@

.PHONY: release
release: ## 10 Build a release
	#goreleaser --release-notes=release_notes.md
	@goreleaser

.PHONY: release-snapshot
release-snapshot: ## 10 Build a snapshot release
	@goreleaser --snapshot --skip-validate --rm-dist

.PHONY: clean
clean:: ## 90 Clean target
	rm -f quicc
	rm -rf dist/

.PHONY: help
help:: ## 99 This help message
	@echo "$(BIN) make(1) targets:"
	@grep -E '^[a-zA-Z\_\-]+:[:]?.*?## [0-9]+ .*$$' $(MAKEFILE_LIST) | \
		sort -n -t '#' -k3,1 | awk '				\
BEGIN { FS = ":[:]?.*?## "; section = 10; };				\
{									\
	newSect = int($$2);						\
	if (section != newSect) {					\
		section = newSect;					\
		printf "\n";						\
	}								\
	sub("^[0-9]+", "",  $$2);					\
	printf "\033[36m%-15s\033[0m %s\n", $$1, $$2;			\
}'
