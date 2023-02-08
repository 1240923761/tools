package main

import (
	"context"
	"errors"
	"time"
	"tools/common/time_wheel"
)

var (
	//generic no matter
	tw = time_wheel.NewTimeWheel[int](context.Background(), 100*time.Millisecond, 10)
)

func init() {
	//k8s.MustInitK8SClient()

	//tw.Start()
	//tw.AfterFunc(5*time.Second, func() {
	//	fmt.Println("hello world")
	//})
	//tw.AfterFunc(10*time.Second, func() {
	//	fmt.Println("hellllllllllllllll")
	//})
	//time.Sleep(100 * time.Second)
}

var (
	errA = errors.New("A error")
	errB = errors.New("B error")
	errC = errors.New("C error")
	errs = []error{errA, errB, errC}
)

func main() {

}
