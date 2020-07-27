package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/martinkunc/wagather/exec"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	w, close, err := exec.NewWebAssembly()
	if err != nil {
		panic(err)
	}
	_ = close
	//defer close()

	if err != nil {
		panic(err)
	}
	c, kcbytes, err := v1Client()
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	w.Configv1client = c
	w.KubeConfigBytes = kcbytes
	result, err := w.Gather()
	if err != nil {
		panic(err)
	}

	fmt.Printf("main: %s \n", result)
}

func v1Client() (*configv1client.ConfigV1Client, []byte, error) {

	kubeconfigBytes, err := ioutil.ReadFile(os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, nil, err
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigBytes)
	if err != nil {
		panic(err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	configv1Client, err := configv1client.NewForConfig(restConfig)

	return configv1Client, kubeconfigBytes, err
}
