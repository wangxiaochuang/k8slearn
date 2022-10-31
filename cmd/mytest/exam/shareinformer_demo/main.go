package shareinformer_demo

import (
	"time"

	"k8s.io/apimachinery/pkg/labels"
	clientgoinformers "k8s.io/client-go/informers"
	"k8s.io/utils/wxc"
)

func init() {
}

func main() {
	informers := clientgoinformers.NewSharedInformerFactory(nil, 10*time.Minute)
	ch := make(chan struct{})
	informers.Start(ch)
	ret, _ := informers.Core().V1().Namespaces().Lister().List(labels.Everything())
	wxc.P(ret)
}
