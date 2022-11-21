package wxc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kr/pretty"
)

func P(objs ...interface{}) {
	if len(objs) == 1 {
		fmt.Printf("%# v\n", pretty.Formatter(objs[0]))
	} else {
		fmt.Printf("%# v\n", pretty.Formatter(objs))
	}
	time.Sleep(time.Hour)
}

func Print(objs ...interface{}) {
	if len(objs) == 1 {
		fmt.Printf("wxc.Print => %# v\n", pretty.Formatter(objs[0]))
	} else {
		fmt.Printf("wxc.Print => %# v\n", pretty.Formatter(objs))
	}
}

func CondP(r *http.Request, objs ...interface{}) {
	if r.URL.Query().Get("wxcdebug") != "true" {
		return
	}
	P(objs...)
}

func CondPrint(r *http.Request, objs ...interface{}) {
	if r.URL.Query().Get("wxcdebug") != "true" {
		return
	}
	Print(objs...)
}
