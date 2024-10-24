run:
	gofumpt -l -w .
	go build -o .
	./triple-s