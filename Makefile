run: build
	./bean-machine -m ~/muzak serve

build:
	go build
	go vet
	go test

clean:
	rm -f bean-machine
