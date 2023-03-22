default: build ;

prepare:
	@go mod tidy
	@mkdir -p bin

build: prepare
	@go build -o ./bin/bima

install:
	mv ./bin/bima /usr/local/bin

uninstall:
	rm -f /usr/local/bin/bima

clean:
	@rm -fr ./bin/