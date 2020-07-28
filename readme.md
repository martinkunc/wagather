## Gathering from Openshift Api using WebAssembly
The goal of this application is to prepare a component, which would could be dynamically replaced during the run of "host" process (Insights Operator).
The host process could be checking defined component url and download the new version when needed and start using it next gathering period.
In this case, the component is WebAssembly, which is executed by hosting process in WebAssembly runtime environment.

In this simple application, [cmd/main.go](/cmd/main.go) loads Kubernetes config and creates client. Then it instantiates wasmer runtime, loads [WebAssembly](/cmd/wasmgather.wasm).
The main then calls through proxy class [exec/exec.go](/exec/exec.go) method gather from inside of WebAssembly.

The idea is similar to IO Gather method. Inside of WebAssembly, method gather from [wasmgather/wasmgather.go](wasmgather/wasmgather.go) calls gatheroperators from inside the same file, which uses Rest Api proxy to gather cluster operators using function `gatherOperators` from same file, very similarly like in IO (when kubernetes api is used.)


### How is the example using Rest Api

For more details why is Openshift client-go not compatible with WebAssembly [See](#client-go-wasm-incomp) 


### Similar solution using JavaScript
The idea is similar like with https://github.com/martinkunc/jsgather. JSGather's component is a Javascript file, which can be in similar fashion be downloaded from Url with
fresh changes.

## What is WebAssembly
Primarily targeted for Web it is a binary with bytecode, which contains compiled code. Many languages have options to compile to WebAssembly. Currently it is supported in modern browsers
so basically besides having JavaScript on a page, I can have .wasm binary, and call it from browser Javascript code. 
For example look here: https://wasmbyexample.dev/examples/hello-world/demo/go/?version=undefined
Because primarily it is used to comlement javascript, lot of implementations is using it from WebPages.
One example how to use it from WebPage is here: https://wasmbyexample.dev/examples/hello-world/hello-world.go.en-us.html
Anyways, because it is just a bytecode, which can be executed by proper runtime, there are more runtime engines, not just in Browsers. It can be started from
NodeJS: https://stackoverflow.com/questions/51403326/how-to-use-webassembly-from-node-js
or many other runtimes (https://github.com/mathetake/gasm, https://wasmer.io/)

## Primary use of WebAssembly in Golang
Naturally I wanted to use Golang capabilities to compile into .wasm. Unfortunatelly today Golang is primarily being used as client side component, and it offers Javascript interop capabilities https://github.com/golang/go/wiki/WebAssembly using (https://golang.org/pkg/syscall/js/) which is marked as Experimental.

## How can be WebAssemblies used from non-javascript world
WebAssembly standard supports interop capabilities, it can Import and Export functions into and from runtime. These functions are today not standarized, and runtimes are importing various functions into WebAssembly runtime. 
For example Golang javascript execution script is importing these functions here: https://github.com/golang/go/blob/master/misc/wasm/wasm_exec.js?source=post_page---------------------------#L248

## Which runtime have I used ?
I have choosen Wasmer (http://wasmer.io) it is runtime written in C and has libraries for Rust, Go and some more languages.

## What does runtime needs to execute Golang code
The Wasmer is trying to export functions required by Golang Javascript execution runtime. And because Golang interface is changing, to compile golang bridge (hack to export required functions) I needed add two functions.
The bridge is based on https://github.com/go-wasm-adapter/go-wasm with neccessary additions for to support Golang 1.14 are in https://github.com/martinkunc/wagather/tree/master/wasm.



## How the application works
How does it work. Golang instantiated Wasmer, the Wasm C execution runtime using its golang library. The WebAssembly is compiled from Golang, loaded by main.go and executed using Wasmer.
Main application is instantiating WebAssembly, calls gather method exported from wasm.
The WebAssembly gather is creatint REST -like wrapper, but actually calls to REST are happening using calling callback function in golang. This way, I can create client with kubeconfig in the host environment. Then result is wrapped back to http response and tries to look like it woudld real HTTP response.

### <a name="client-go-wasm-incomp"></a> Why is Openshift client-go not compatible with WebAssembly
Compilation against wasm target, compared for example with compilation with x86 target is using different golang packages by conditional compilation.
client is then unable to find implmentation of terminal, which is implemented per platform, but not for wasm platform.
- runtime package requires x/crypto/terminal, which doesn't exists for wasm platform and importing runtime is panicking. Second issue I found was runtime error: index out of range [0] with length 0
- creates too big .wasm files. The golang 1.13 cannot compile Webassemblies larger then 30 MB. (see https://github.com/golang/go/issues/34395). Without client-go, the wasm is smaller and can be loaded.



### There are following limitations.
Web assembly is compiled using golang 1.13, because in 1.14 the bridge hack would still need to be changed to support new constants.


Another problem is that WASMI interface, for executing WebAssembly on server is not stabilised yet, so Golang implementation is very Nodejs specific and some Api was changed in v.1.14.

To overcome this I decided to write wrapper around REST client with similar structure as kubernetes client. Another problem iss that client-go REST Api client is not mockable, so I had to write a bit more code.

WebAssembly code is in wasmgather folder. It requires to be hosted in main package. Currently it is not separated from main repo with standalone vendors, probably it could.
It generating web assembly wasmgather.wasm, which has today about 7MB. 

Golang Wasm compiler is building very large binaries. Tinygo is (much) better, but it doesn't support reflect package (I tried using client-go), maybe using current REST wrapper it would work better. This is also causing long laoding times of WebAssembly. Now it loads in 33 seconds.



## Problems with compiling with tinygo
WebAssembly produced by tinygo is about 200K.
The interface for communication from wasm host to web assembly is not standarized, tinygo uses different interface functions that the ones golang uses, which results in failing on Imports. The golang imports are implemented by imports.go, so potential change there would allow using tinygo.

Also it is not possible to use K8s client-go, because it uses reflect, which is not fully implemented in tinygo.


## Next steps
Next steps would be investigate how to automate building REST Proxy classes. I can see a lot of improvements using kubernetes code generators for example.
