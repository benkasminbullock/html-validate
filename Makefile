all: validate

INSTALLED=/home/ben/bin/validate

install: validate
	if [ -f $(INSTALLED) ]; then chmod 0755 $(INSTALLED); fi
	cp validate $(INSTALLED)
	chmod 0555 $(INSTALLED)

validate: validate.go tag.go
	go build validate.go tag.go

test: validate run-tests.pl
	prove ./run-tests.pl

clean:
	purge -r
	-rm -f validate
