.PHONY: all deps test-api test-ui test-all clean

all: deps test-all

deps:
	go mod download
	go mod tidy

test-api:
	./run-tests.sh -run=TestAPI

test-ui:
	./run-tests.sh -run=TestUI

test-all:
	./run-tests.sh

clean:
	rm -rf evidence/
