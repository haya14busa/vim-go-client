# Prefer go commands for basic tasks like `build`, `test`, etc...

.PHONY: lintdeps lint

lintdeps:
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
	go get github.com/client9/misspell/cmd/misspell

lint: lintdeps
	! gofmt -d -s . | grep '^' # exit 1 if any output given
	! golint ./... | grep '^'
	go vet ./...
	errcheck -asserts -ignoretests -ignore 'Close|Start|Serve|Remove'
	misspell -error **/*.go **/*.md
