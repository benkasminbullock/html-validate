all: validate
#	./validate t/books2.html
#	./validate t/soft-keyboard.html
	./validate t/mr-old.html

INSTALLED=/home/ben/bin/validate

install: validate
	if [ -f $(INSTALLED) ]; then chmod 0755 $(INSTALLED); fi
	cp validate $(INSTALLED)
	chmod 0555 $(INSTALLED)

validate: validate.go tag.go
	go build validate.go tag.go

clean:
	-rm -f validate
