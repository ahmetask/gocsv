package main

import (
	"errors"
	"fmt"
	"github.com/ahmetask/gocsv"
	"reflect"
	"sync"
	"time"
)

type A struct {
	X string
	T int
}
type Model struct {
	A
}

func convertField(v string, k reflect.Kind, t reflect.Type) (interface{}, error) {
	return nil, errors.New("error data:" + v + "-k:" + k.String() + "-t:" + t.String())
}

func main() {
	reader, err := gocsv.NewReader(gocsv.ReaderConfig{
		FilePath: "./example/json/json.txt",
	}, Model{}, convertField)
	if err != nil {
		fmt.Println(err)
		return
	}

	readChannel := reader.Read()

	fmt.Println("start:" + time.Now().String())
	wg := &sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(w *sync.WaitGroup) {
			for r := range readChannel {
				if r.Exist() {
					if v, ok := r.Value().(*Model); ok {
						v.T = 1
					}
				} else {
					fmt.Println(r.Err())
				}
			}
			w.Done()
		}(wg)
	}

	wg.Wait()
	_ = reader.Close()
	fmt.Println("end:" + time.Now().String())

}
