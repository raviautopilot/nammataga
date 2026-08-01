.PHONY: all deps test-api test-ui test-all clean

all: deps test-all

deps:
	go mod download
	go mod tidy

test-api:
	./run-api-tests.sh

test-ui:
	./run-ui-tests.sh

test-all:
	./run-tests.sh

clean:
	rm -rf evidence/
