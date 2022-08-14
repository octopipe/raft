BUF_VERSION:=1.1.0

generate:
	docker run -v $$(pwd):/src -w /src --rm bufbuild/buf:$(BUF_VERSION) generate
