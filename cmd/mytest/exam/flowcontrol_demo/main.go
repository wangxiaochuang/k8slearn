package flowcontrol_demo

import (
	"context"
	"fmt"

	flowcontrol "k8s.io/client-go/util/flowcontrol"
	"k8s.io/utils/wxc"
)

func init() {
}

func main() {
	f := flowcontrol.NewTokenBucketRateLimiter(2, 2)
	for {
		f.Wait(context.Background())
		fmt.Println("accept one again")
	}
	wxc.P(f)
}
