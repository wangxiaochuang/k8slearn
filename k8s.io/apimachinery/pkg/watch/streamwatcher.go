package watch

import (
	"fmt"
	"io"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
)

type Decoder interface {
	Decode() (action EventType, object runtime.Object, err error)
	Close()
}

type Reporter interface {
	AsObject(err error) runtime.Object
}

type StreamWatcher struct {
	sync.Mutex
	source   Decoder
	reporter Reporter
	result   chan Event
	done     chan struct{}
}

func NewStreamWatcher(d Decoder, r Reporter) *StreamWatcher {
	sw := &StreamWatcher{
		source:   d,
		reporter: r,
		result:   make(chan Event),
		done:     make(chan struct{}),
	}
	go sw.receive()
	return sw
}

func (sw *StreamWatcher) ResultChan() <-chan Event {
	return sw.result
}

func (sw *StreamWatcher) Stop() {
	sw.Lock()
	defer sw.Unlock()
	select {
	case <-sw.done:
	default:
		close(sw.done)
		sw.source.Close()
	}
}

func (sw *StreamWatcher) receive() {
	defer utilruntime.HandleCrash()
	defer close(sw.result)
	defer sw.Stop()
	for {
		action, obj, err := sw.source.Decode()
		if err != nil {
			switch err {
			case io.EOF:
				// watch closed normally
			case io.ErrUnexpectedEOF:
				klog.V(1).Infof("Unexpected EOF during watch stream event decoding: %v", err)
			default:
				if net.IsProbableEOF(err) || net.IsTimeout(err) {
					klog.V(5).Infof("Unable to decode an event from the watch stream: %v", err)
				} else {
					select {
					case <-sw.done:
					case sw.result <- Event{
						Type:   Error,
						Object: sw.reporter.AsObject(fmt.Errorf("unable to decode an event from the watch stream: %v", err)),
					}:
					}
				}
			}
			return
		}
		select {
		case <-sw.done:
			return
		case sw.result <- Event{
			Type:   action,
			Object: obj,
		}:
		}
	}
}
