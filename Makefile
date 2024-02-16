
build-docker:
	docker build -t pow .

run-compose:
	docker-compose up -d --force-recreate

lint::
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 -v run ./...

test::
	go test -coverpkg=./... -race -coverprofile=cover.out.tmp -covermode atomic -v ./...
	cat cover.out.tmp > coverage.txt # strip out generated go-connect files
	go tool cover -func coverage.txt
	rm cover.out.tmp coverage.txt