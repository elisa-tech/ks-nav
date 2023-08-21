EXECUTABLE=navweb
INSTALL_DIR=/usr/bin/
GOBIN ?= ~/go/bin

all: main.go data.go
	go build -o $(EXECUTABLE)

data.go: data
	${GOBIN}/go-bindata -o data.go data/templates/ data/configs

install: $(EXECUTABLE)
	cp $(EXECUTABLE)  $(INSTALL_DIR)
