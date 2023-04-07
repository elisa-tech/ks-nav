all:	nav.go
	go build -ldflags="-r ."

upx:	nav
	upx nav

test:
	go test -cover -v
	go test ./config -v -cover

code/check:
	@gofmt -l `find . -type f -name '*.go' -not -path "./vendor/*"`

code/fmt:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`
