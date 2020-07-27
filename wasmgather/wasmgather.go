// +
// build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"syscall/js"
	//configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createString(this js.Value, args []js.Value) interface{} {
	if args == nil {
		logf("createString input is nil")
	}

	b := args[0].String()
	logf("createString: b = %s\n len(b) %d", b, len(b))

	v := js.Global().Get("Uint8Array").New(len(b)) // no
	js.CopyBytesToJS(v, []byte(b))
	return v
}

// readStringVal reads js.Val from js runtime to Go
func readStringVal(s js.Value) string {
	st := s.String()
	l := s.Length()
	logf("readStringVal: b = %s length %d\n", st, l)
	// read bytes
	buf := make([]byte, l, l)
	js.CopyBytesToGo(buf, s)
	return string(buf)
}

// writeStringVal writes js.Val from Go to js runtime
func writeStringVal(b string) interface{} {
	r := []byte(b)
	logf("writeStringVal with bytes: r = %s\n len(r) %d", r, len(r))

	// write bytes
	v := js.Global().Get("Uint8Array").New(len(r))
	js.CopyBytesToJS(v, r)
	return v
}

type CustomHttpClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

type CustomHttpClient struct{}

func (c *CustomHttpClient) Do(r *http.Request) (*http.Response, error) {
	logf("Inside CustomHttpClient.Do")
	logf("url %s %s", r.URL, r.Method)

	preq, err := PackReq(r)
	if err != nil {
		return nil, err
	}
	pres := httpCallProxy(preq)
	resp, err := UnpackRes(pres)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func gather(this js.Value, args []js.Value) interface{} {
	logf("inside gather")
	kcVal := args[0]
	logf("kc : b = %s\n", kcVal.String())

	kc := readStringVal(kcVal)
	_ = kc
	rs := map[string]string{}

	cl := &CustomHttpClient{}
	c := createv1Client(cl)

	r, err := gatherOperators(c)
	logf("After gatherOperators")
	if err != nil {
		log.Printf("error %s", err)
		return writeStringVal("error: " + err.Error()) // err
	}

	merge(rs, r)
	logf("After gather merge")
	bts, err := json.Marshal(rs)
	if err != nil {
		log.Printf("error in marshal %s", err)
		return writeStringVal("error: " + err.Error()) // err
	}

	res := string(bts)

	logf("Returning: %s", res)
	return writeStringVal(res)
}

func gatherOperators(configClient *configv1clientProxy) (map[string]string, error) {
	//c := configv1client.ConfigV1Interface.ClusterOperators().List(metav1.ListOptions{} ) ClusterOperators().C

	l, err := configClient.ClusterOperators().List(ListOptionsProxy{})
	if err != nil {
		return nil, fmt.Errorf("gatherOperators ClusterOperators.List error: %w", err)
	}
	bts, err := json.Marshal(l)
	if err != nil {
		return nil, fmt.Errorf("gatherOperators marshal error: %w", err)
	}
	r := map[string]string{"clusterOperators": string(bts)}
	return r, nil
}

func merge(d map[string]string, s map[string]string) {
	for k, v := range s {
		d[k] = v
	}
}

func httpCallProxy(u string) string {
	//uVal := createString( js.Value u)
	res := js.Global().Get("httpCall").Invoke(u)
	s := readStringVal(res)
	logf("httpCallProxy res %s", s)
	return s
}

func logf(format string, v ...interface{}) {
	if LOGENABLED {
		log.Printf(format, v...)
	}
}

var LOGENABLED = false

func main() {
	ch := make(chan bool)
	logf("inside wasm main")
	// register functions
	js.Global().Set("createString", js.FuncOf(createString))
	js.Global().Set("gather", js.FuncOf(gather))

	<-ch
}
