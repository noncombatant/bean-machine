build: test
	go vet
	staticcheck -checks all
	go build

run: build
	./bean-machine -m ~/muzak catalog serve

clean:
	@rm -f bean-machine
	@rm -f coverage.out coverage.html

test:
	go test -coverprofile="coverage.out"
	go tool cover -html="coverage.out"
	gosloc
