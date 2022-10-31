package main

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/utils/wxc"
)

func init() {
}

type listen struct{}

func (listen) Enqueue() {
	fmt.Printf("changed")
}
func main() {
	c, _ := dynamiccertificates.NewDynamicCAContentFromFile("client-ca-bundle", "testdata/tmp.crt")
	c.AddListener(listen{})
	go c.Run(context.Background(), 4)
	for {
		select {
		case <-time.NewTicker(5 * time.Second).C:
			a := c.CurrentCABundleContent()[:9]
			fmt.Printf("%02X\n", a)
		}
	}
	wxc.P(c)
}
