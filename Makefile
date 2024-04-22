build:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/dependabot ./cmd/dependabot/

run:
	./bin/dependabot config preview -o github
