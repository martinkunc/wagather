// +build !js

package exec

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/martinkunc/wagather/wasm"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewWebAssembly() (*WebAssembly, func(), error) {
	n := time.Now()
	b, dealloc, err := wasm.BridgeFromFile("test", "wasmgather.wasm", nil)
	if err != nil {
		return nil, nil, err
	}
	// 12s
	log.Println(time.Now().Sub(n))

	w := &WebAssembly{}
	err = b.SetFunc("httpCall", w.httpCall(b))
	if err != nil {
		return nil, nil, err
	}

	init := make(chan error)
	ctx, cancF := context.WithCancel(context.Background())
	//defer cancF()
	_ = cancF
	go b.Run(ctx, init)
	err = <-init
	if err != nil {
		return nil, nil, err
	}

	w.bridge = b
	return w, dealloc, nil
}

type WebAssembly struct {
	KubeConfigBytes []byte
	Configv1client  *configv1client.ConfigV1Client
	bridge          *wasm.Bridge
}

func (w *WebAssembly) createString(s string) (interface{}, error) {
	return w.bridge.CallFunc("createString", []interface{}{s})
}

// readString reads js.Val value from WA runtime and converts it to string
func (w *WebAssembly) readString(jsVal interface{}) (string, error) {
	nb, err := wasm.Bytes(jsVal)
	if err != nil {
		return "", fmt.Errorf("readString Bytes error: %w", err)
	}
	sb := strings.Builder{}
	for i := 0; i < len(nb); i++ {
		sb.WriteByte(nb[i])
	}
	return sb.String(), nil
}

func clientFromBytes(b []byte) (*configv1client.ConfigV1Client, *rest.Config, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		panic(err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	configv1Client, err := configv1client.NewForConfig(restConfig)

	return configv1Client, restConfig, err
}

// httpCall is the callback from WebAssembly to go
func (w *WebAssembly) httpCall(b *wasm.Bridge) wasm.Func {
	return func(args []interface{}) (i interface{}, e error) {

		log.Println("In Go", args)
		if len(args) != 1 {
			log.Println("not enough arguments for httpCall. got: %d", len(args))
			return nil, fmt.Errorf("not enough args")
		}
		req, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("args[0] is not string")
		}
		r, err := UnpackReq(req)
		if err != nil {
			return nil, fmt.Errorf("unpackReq error: %w", err)
		}

		configv1, rc, err := clientFromBytes(w.KubeConfigBytes)
		if err != nil {
			return nil, fmt.Errorf("clientfrombytes error: %w", err)
		}
		log.Printf("Req url %s Api Url %s", r.URL.String(), rc.APIPath)
		//resource := "clusteroperators"

		su := strings.Split(r.URL.String(), "/")
		if len(su) < 2 {
			return nil, fmt.Errorf("cannot find resource from url: %s", r.URL.String())
		}
		// group / version / resource
		resource := su[2]
		rr, err := configv1.RESTClient().Verb(r.Method).Resource(resource).DoRaw()
		status := 200
		if err != nil {
			status = 500
			log.Printf("Api error: %s", err)
		}
		log.Printf("Status  %d", status)
		res, err := PackRes(r.URL.String(), r.Method, rr, status)
		if err != nil {
			return nil, fmt.Errorf("PackRes error: %w", err)
		}

		sVal, err := w.createString(res)
		if err != nil {
			return "", fmt.Errorf("Gather createString error: %w", err)
		}

		return sVal, nil

	}
}

func (w *WebAssembly) Gather() (string, error) {

	kubeVal, err := w.createString(string(w.KubeConfigBytes))
	if err != nil {
		return "", err
	}
	resVal, err := w.bridge.CallFunc("gather", []interface{}{kubeVal})
	if err != nil {
		return "", fmt.Errorf("Gather callFunc error: %w", err)
	}
	result, err := w.readString(resVal)
	if err != nil {
		return "", fmt.Errorf("Gather readString error: %w", err)
	}

	return result, nil
}
