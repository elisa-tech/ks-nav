
all:	nav.go
	go build -ldflags="-r ."

cgo:    nav.go dot_parser/libdotparser.so libdotparser.so
	go build -ldflags="-r ." -tags CGO

upx:	nav
	upx nav

test:
	go test -cover -v
	go test ./config -v -cover

code/check:
	@gofmt -l `find . -type f -name '*.go' -not -path "./vendor/*"`

code/fmt:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

dot_parser/libdotparser.so: dot_parser/dot.l dot_parser/dot.y
	$(MAKE) -C dot_parser/ libdotparser.so

libdotparser.so: dot_parser/libdotparser.so
	ln -s dot_parser/libdotparser.so libdotparser.so

clean: 
	$(MAKE) -C dot_parser/ clean
	rm -f nav libdotparser.so
