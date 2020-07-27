# build-wasm:
# 	cp hack/vendor/golang.org/x/crypto/ssh/terminal/util_wasm.go vendor/golang.org/x/crypto/ssh/terminal/
# 	GOFLAGS=-mod=vendor GOOS=js GOARCH=wasm go build -o wasmgather.wasm wasmgather/*.go

build-wasm:
	docker run --rm -v "${PWD}":/go/src/github.com/martinkunc/wagather -w /go/src/github.com/martinkunc/wagather golang:1.13 bash -c "GOFLAGS=-mod=vendor GOOS=js GOARCH=wasm go build -o wasmgather.wasm wasmgather/*.go"
	test $$? -eq 0 && cp wasmgather.wasm cmd/



build:
	go build -o wagather ./cmd/main.go

