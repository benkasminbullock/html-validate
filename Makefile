all: validate
	./validate t/books2.html

validate: validate.go
	go build validate.go

clean:
	-rm -f validate
