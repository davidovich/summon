SUMMONER_NAME := {{ .SummonerName }}

ASSETS := $(shell find $(SUMMONER_NAME)/assets)

all: bin/$(SUMMONER_NAME)

bin/$(SUMMONER_NAME): $(SUMMONER_NAME)/$(SUMMONER_NAME).go $(ASSETS)
	go build -o $@ $<

.PHONY: clean
clean:
	rm -rf bin
