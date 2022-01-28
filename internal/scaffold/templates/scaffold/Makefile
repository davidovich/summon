SUMMONER_NAME := {{ .SummonerName }}

ASSETS := $(shell find $(SUMMONER_NAME)/assets)

all: bin/$(SUMMONER_NAME)

bin/$(SUMMONER_NAME): $(SUMMONER_NAME)/$(SUMMONER_NAME).go $(ASSETS)
	go build -o $@ $<
	@echo testing created $(SUMMONER_NAME)
	go run $< --help

.PHONY: clean
clean:
	rm -rf bin
