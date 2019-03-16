SHELL=/bin/bash

HAS_PACKR2	:= $(shell command -v packr2)

SUMMONER_NAME := {{ .SummonerName }}

SUMMON_DEPS = $(shell find $(SUMMONER_NAME))
ASSETS := $(shell find assets)

all: $(SUMMONER_NAME)/packrd/packed-packr.go bin/$(SUMMONER_NAME)

bin/$(SUMMONER_NAME): $(SUMMON_DEPS)
	@mkdir -p bin
	go build -o $@ $(@F)/summon.go

$(SUMMONER_NAME)/packrd/packed-packr.go: $(ASSETS)
ifndef HAS_PACKR2
	$(error packr2 was not found on path and is needed to make assets)
endif
	cd $(SUMMONER_NAME) && GO111MODULE=on packr2

.PHONY: clean
clean:
	cd $(SUMMONER_NAME) && packr2 clean