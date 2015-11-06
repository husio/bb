build:
	GO15VENDOREXPERIMENT=1 go build -o bb

dev: build
	DEV=1 ./bb

run: build
	./bb

clean:
	@@rm ./bb 2> /dev/null


.PHONY: build dev run clean
