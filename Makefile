all:	nav.go
	go build

aarch64: nav.go
	GOARCH="arm64" GOOS="linux" go build

amd64: nav.go
	GOARCH="amd64" GOOS="linux" go build

upx:	nav
	upx nav

test:
	go test -cover
	go test ./config -v -cover

code/check:
	@gofmt -l `find . -type f -name '*.go' -not -path "./vendor/*"`

code/fmt:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`
