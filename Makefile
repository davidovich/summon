SHELL=/bin/bash

export GO111MODULE := on

SCAFFOLD_BIN := bin/scaffold
SRCS = $(shell go list -buildvcs=false -f '{{ $$dir := .Dir}}{{range .GoFiles}}{{ printf "%s/%s\n" $$dir . }}{{end}}' $(1)/... github.com/davidovich/summon/...)

DOC_REPO_NAME := davidovich.github.io
DOC_REPO := git@github.com:davidovich/$(DOC_REPO_NAME).git
SUMMON_BADGE_JSON_FILE := $(DOC_REPO_NAME)/shields/summon/summon.json

ASSETS := $(shell find internal/scaffold/templates/scaffold)

all: test $(SCAFFOLD_BIN)

.PHONY: bin
bin: $(SCAFFOLD_BIN)

.PHONY: examples
examples: bin/cmd-proxy

bin/cmd-proxy: $(call SRCS,github.com/davidovich/summon/examples/cmd-proxy) $(shell find examples/cmd-proxy/assets)
	go build -o $@ github.com/davidovich/summon/examples/$(@F)


.PHONY: $(SCAFFOLD_BIN)
$(SCAFFOLD_BIN): $(ASSETS) $(call SRCS,github.com/davidovich/summon/scaffold)
	go build -o $@ $(@F)/$(@F).go

COVERAGE := build/coverage/coverage.out
COVERAGE_PERCENT_FILE := build/coverage/percent.txt
HTML_COVERAGE := build/coverage/index.html

.PHONY: test
test: clean-coverage output-coverage

.PHONY: clean-coverage
clean-coverage:
	rm -f $(COVERAGE)

.PHONY: output-coverage
output-coverage: $(COVERAGE) $(HTML_COVERAGE)
	go tool cover -func=$<

$(COVERAGE_PERCENT_FILE): $(COVERAGE)
	go tool cover -func=$< | sed -n 's/total:[[:space:]]*(statements)[[:space:]]*\([0-9.]*\)%/\1/gw $@'

$(HTML_COVERAGE): $(COVERAGE)
	@mkdir -p $(@D)
	go tool cover -html=$< -o $@

$(COVERAGE):
	@mkdir -p $(@D)
	go test -buildvcs=false ./... -timeout 30s --coverpkg=./... -coverprofile $@ -v

.PHONY: update-coverage-badge
update-coverage-badge: $(COVERAGE_PERCENT_FILE)
ifneq ("$(CIRCLE_BRANCH)","master")
	@echo
	@echo "On branch $(CIRCLE_BRANCH), not publishing this branch's $$(cat $(COVERAGE_PERCENT_FILE))% total coverage."
else
	rm -rf /tmp/$(DOC_REPO_NAME)
	git -C /tmp clone $(DOC_REPO)
	cd /tmp/$(DOC_REPO_NAME) && \
	go run github.com/davidovich/summon-example-assets@3c2e66d7 shields.io/coverage.json.gotmpl --json "{\"Coverage\": \"$$(cat $(COVERAGE_PERCENT_FILE))\"}" -o- > /tmp/$(SUMMON_BADGE_JSON_FILE)
	git -C /tmp/$(DOC_REPO_NAME) diff-index --quiet HEAD || git -C /tmp/$(DOC_REPO_NAME) -c user.email=automation@davidovich.com -c user.name=automation commit -am "automated coverage commit of $$(cat $(COVERAGE_PERCENT_FILE)) %" || true
	git -C /tmp/$(DOC_REPO_NAME) push
endif

.PHONY: clean
clean:
	rm -rf bin build
