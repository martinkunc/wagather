package main

import (
	"log"
)

func createv1Client(c CustomHttpClientInterface) *configv1clientProxy {
	return &configv1clientProxy{
		HttpClient:       c,
		client:           &RestClientProxy{c},
		groupWithVersion: "config.openshift.io/v1",
	}
}

type configv1clientProxy struct {
	HttpClient       CustomHttpClientInterface
	client           *RestClientProxy
	groupWithVersion string
}

func (c *configv1clientProxy) ClusterOperators() *ClusterOperatorInterfaceProxy {
	return &ClusterOperatorInterfaceProxy{configv1clientProxy: c}
}

type ClusterOperatorInterfaceProxy struct {
	*configv1clientProxy
}

// ClusterOperatorInterfaceProxy
func (c *ClusterOperatorInterfaceProxy) List(opts ListOptionsProxy) (string, error) {
	var result string
	log.Println("Before executing client.get")

	err := c.client.Get().
		Resource("clusteroperators").
		VersionedParams(&opts, c.groupWithVersion).
		Do().
		Into(&result)
	if err != nil {
		log.Printf("err in client get %s\n", err.Error())
		return "", err
	}
	return result, err

	// err = c.client.Get().
	// 	Resource("clusteroperators").
	// 	VersionedParams(&opts, scheme.ParameterCodec).
	// 	Timeout(timeout).
	// 	Do().
	// 	Into(result)
}

type ListOptionsProxy struct{}
