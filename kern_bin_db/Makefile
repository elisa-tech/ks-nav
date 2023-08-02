EXECUTABLE=nav-db-filler
INSTALL_DIR=/usr/bin/

all:	main.go
	go build  -o $(EXECUTABLE)

aarch64: main.go
	GOARCH="arm64" GOOS="linux" go build -o $(EXECUTABLE)

amd64: main.go
	GOARCH="amd64" GOOS="linux" go build -o $(EXECUTABLE)

upx:	$(EXECUTABLE)
	upx $(EXECUTABLE)

test:
	go test

install:
	cp $(EXECUTABLE) $(INSTALL_DIR)
