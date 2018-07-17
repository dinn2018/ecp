PACKAGE = github.com/dinn2018/ecp
export GOPATH = $(CURDIR)/.build
SRC_BASE = $(GOPATH)/src/$(PACKAGE)

.PHONY: ecp all

faucet: |$(SRC_BASE)
	@cd $(SRC_BASE) && go build -i -o bin/ecp -v ./cmd/*
	@echo "use bin/ecp start"

$(SRC_BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

all: ecp 