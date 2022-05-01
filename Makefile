build:
	go build
	go vet
	go test

run: build
	./bean-machine -m ~/muzak catalog serve

clean:
	rm -f bean-machine
