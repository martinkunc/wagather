Gathering from Openshift using WebAssembly
How does it work. Golang instantiated Wasmer, the Wasm C execution runtime using its golang library. The WebAssembly is compiled from Golang, loaded by main.go and executed using Wasmer.

There are following limitations.
Web assembly is compiled using golang 1.13, because in 1.14 there is a bug: https://github.com/golang/go/issues/40366
Openshift/Kubernetes client-go is not compatible with WebAssembly 
- runtime package requires x/crypto/terminal, which doesn't exists for wasm platform and importing runtime is panicking. Second issue I found was runtime error: index out of range [0] with length 0
- creates too big .wasm files. The golang 1.13 cannot compile Webassemblies larger then 30 MB. (Not a problem now)

Another problem is that WASMI interface, for executing WebAssembly on server is not stabilised yet, so Golang implementation is very Nodejs specific and some Api was changed in v.1.14.

To overcome this I decided to write wrapper around REST client with similar structure as kubernetes client. Another problem iss that client-go REST Api client is not mockable, so I had to write a bit more code.

WebAssembly code is in wasmgather folder. It requires to be hosted in main package. Currently it is not separated from main repo with standalone vendors, probably it could.
It generating web assembly wasmgather.wasm, which has today about 7MB. 

Golang Wasm compiler is building very large binaries. Tinygo is (much) better, but it doesn't support reflect package (I tried using client-go), maybe using current REST wrapper it would work better. This is also causing long laoding times of WebAssembly. Now it loads in 33 seconds.

Main application is instantiating WebAssembly, calls gather method exported from wasm.
The WebAssembly gather is creatint REST -like wrapper, but actually calls to REST are happening using calling callback function in golang. This way, I can create client with kubeconfig in the host environment. Then result is wrapped back to http response and tries to look like it woudld real HTTP response.

Next steps would be investigate how to automate building REST Proxy classes. I can see a lot of improvements using kubernetes code generators for example.





##Why not tinygo yet
WebAssembly produced by tinygo is about 200K.
The interface for communication from wasm host to web assembly is not standarized, tinygo uses different interface functions that the ones golang uses, which results in failing on Imports.
The golang imports are implemented by imports.go, so potential change there would allow using tinygo.

docker run --rm -v "${PWD}":/go/src/github.com/martinkunc/wagather -w /go/src/github.com/martinkunc/wagather golang:1.13 bash -c "GOFLAGS=-mod=vendor GOOS=js GOARCH=wasm go build -o wasmgather.wasm wasmgather/wasmgather.go"

docker run --rm -it -v "${PWD}":/go/src/github.com/martinkunc/wagather -e "GOPATH=/go" -w /go/src/github.com/martinkunc/wagather tinygo/tinygo:0.13.1 bash -c

tinygo build -o /go/src/github.com/myuser/myrepo/wasm.wasm -target wasm --no-debug /go/src/github.com/myuser/myrepo/wasm-main.go

old build

build-wasm:
	cp hack/vendor/golang.org/x/crypto/ssh/terminal/util_wasm.go vendor/golang.org/x/crypto/ssh/terminal/
	docker run --rm -v "${PWD}":/go/src/github.com/martinkunc/wagather -w /go/src/github.com/martinkunc/wagather golang:1.13 bash -c "GOFLAGS=-mod=vendor GOOS=js GOARCH=wasm go build -o wasmgather.wasm wasmgather/*.go"
