PROG=	dcc
DEST=	$(HOME)/bin

.PHONY: $(PROG) all clean install

all:		$(PROG)
$(PROG):;	@go build -o $@ && go vet
clean:;		@rm -f $(PROG)
install: $(PROG); install -c -m 555 $(PROG) $(DEST)/$(PROG)
