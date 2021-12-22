package main

import (
	"fmt"
)

var (
	ErrorInvalidArgument = fmt.Errorf("invalid argument")
)

type Service interface {
	Init(args map[string]interface{}) error
	Run()
	Rehash()
	Shutdown()
}
