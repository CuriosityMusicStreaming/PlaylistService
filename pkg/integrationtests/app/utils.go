package app

import (
	"fmt"
	"reflect"
)

func assertNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func assertEqual(expect, got interface{}) {
	equal := reflect.DeepEqual(expect, got)
	if !equal {
		panic(fmt.Sprintf("expected equal params expect: %s got %s", expect, got))
	}
}
