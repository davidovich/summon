SHELL=/bin/bash

HAS_PACKR2				:= $(shell command -v packr2)
HAS_GOBIN				:= $(shell command -v gobin)
HAS_GOCOVERUTIL			:= $(shell command -v gocoverutil)

ifndef HAS_GOBIN
$(shell GO111MODULE=off go get -u github.com/myitcv/gobin 2>/dev/null)
endif

export GO111MODULE := on

SCAFFOLD_BIN := bin/scaffold
PACKR_FILE := pkg/scaffold/scaffold-packr.go
COVERAGE_PERCENT_FILE := $(CURDIR)/build/coverage-percent.txt

DOC_REPO_NAME := davidovich.github.io
DOC_REPO := git@github.com:davidovich/$(DOC_REPO_NAME).git
SUMMON_BADGE_JSON_FILE := $(DOC_REPO_NAME)/shields/summon/summon.json

ASSETS := $(shell find templates/scaffold)

all: $(PACKR_FILE) test $(SCAFFOLD_BIN)

.PHONY: bin
bin: $(PACKR_FILE) $(SCAFFOLD_BIN)

.PHONY: $(SCAFFOLD_BIN)
$(SCAFFOLD_BIN): $(PACKR_FILE)
	@mkdir -p bin
	go build -o $@ $(@F)/$(@F).go

$(PACKR_FILE): $(ASSETS)
ifndef HAS_PACKR2
	gobin github.com/gobuffalo/packr/v2/packr2
endif
	packr2

GO_TESTS := pkg internal cmd
COVERAGE := $(foreach t,$(GO_TESTS),build/coverage/report/$(t))
MERGED_COVERAGE := build/coverage/report/cover.merged.out
HTML_COVERAGE := build/coverage/html/index.html

.PHONY: test
test: output-coverage

.PHONY: output-coverage
output-coverage: $(MERGED_COVERAGE) $(HTML_COVERAGE)
	go tool cover -func=$< | sed -e 's/total:[[:space:]]*(statements)[[:space:]]*\([0-9.]*\)%/\1/gw $(COVERAGE_PERCENT_FILE)'

$(MERGED_COVERAGE): $(COVERAGE)
ifndef HAS_GOCOVERUTIL
	gobin github.com/AlekSi/gocoverutil@v0.2.0
endif
	$(call msg,Generating merged coverage report)
	gocoverutil -coverprofile=$@ merge $^

$(HTML_COVERAGE): $(MERGED_COVERAGE)
	@mkdir -p $(@D)
	go tool cover -html=$< -o $@

.PHONY: $(COVERAGE)
$(COVERAGE):
	@mkdir -p $(@D)
	$(call msg,--> Testing $(@F)...)
	go test ./$(@F)/... --cover -coverprofile $@ -v

.PHONY: update-coverage-badge
update-coverage-badge:
	rm -rf /tmp/$(DOC_REPO_NAME)
	git -C /tmp clone $(DOC_REPO)
	cd /tmp/$(DOC_REPO_NAME) && \
	go run github.com/davidovich/summon-example-assets/summon shields.io/coverage.json.gotmpl --json "{\"Coverage\": \"$$(cat $(COVERAGE_PERCENT_FILE))\"}" -o- > /tmp/$(SUMMON_BADGE_JSON_FILE)
	git -C /tmp/$(DOC_REPO_NAME) diff-index --quiet HEAD || git -C /tmp/$(DOC_REPO_NAME) -c user.email=automation@davidovich.com -c user.name=automation commit -am "automated coverage commit of $$(cat $(COVERAGE_PERCENT_FILE)) %" || true
	git -C /tmp/$(DOC_REPO_NAME) push

.PHONY: clean
clean:
	packr2 clean
	rm -r bin build
