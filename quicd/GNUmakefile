BIN=quicd

default:: help

dev:: ## 10 Build a local build
	go build -o quicd

all:: quicd_illumos ## 10 Build all targets

illumos:: ## 10 Build for Illumos
	env GOOS=solaris GOARCH=amd64 go build -o $@

.PHONY: ssl-keys
ssl-keys:: ## 10 Create SSL keys
	openssl req -subj '/CN=quicd.example.com/O=Joyent Inc./C=US' -new -newkey rsa:2048 -days 3650 -nodes -x509 -sha256 -keyout quicd.key -out quicd.pem

.PHONY: release
release: ## 10 Build a release
	#goreleaser --release-notes=release_notes.md
	@goreleaser

.PHONY: release-snapshot
release-snapshot: ## 10 Build a snapshot release
	@goreleaser --snapshot --skip-validate --rm-dist

.PHONY: clean
clean:: ## 90 Clean target
	rm -f quicd
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
	printf "\033[36m%-20s\033[0m %s\n", $$1, $$2;			\
}'
