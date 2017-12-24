dest?=$(HOME)/bin
p=dcc
.PHONY: all clean install
all:; @go build -o $p && go vet
clean:; @rm -f $p
install: all; install -c -m 555 $p $(dest)
