package wxc

import (
	"fmt"

	"github.com/kr/pretty"
)

func P(objs ...interface{}) {
	if len(objs) == 1 {
		fmt.Printf("%# v\n", pretty.Formatter(objs[0]))
	} else {
		fmt.Printf("%# v\n", pretty.Formatter(objs))
	}
}
