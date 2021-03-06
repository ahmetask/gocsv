package main

import (
	"errors"
	"fmt"
	"github.com/ahmetask/gocsv"
	"reflect"
	"sync"
	"time"
)

type MyStruct2 struct {
	Y string
}

type MyStruct1 struct {
	X string
	T int
}

// Model add your structures in the order found in the csv. Json text can not contains your separator text and "\n"
type Model struct {
	MyStruct1
	MyStruct2
}

func convertField(v string, k reflect.Kind, t reflect.Type) (interface{}, error) {
	return nil, errors.New(fmt.Sprintf("convertError data:%v kind:%v type:%v", v, k.String(), t.String()))
}

func main() {
	reader, err := gocsv.NewReader(gocsv.ReaderConfig{
		FilePath:        "./example/json/json.csv",
		Format:          Model{},
		ConvertFunction: convertField,
	})
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
				x := r.OrElse(func(err error) interface{} {
					return err
				})

				fmt.Println(x)
			}
			w.Done()
		}(wg)
	}

	wg.Wait()
	_ = reader.Close()
	fmt.Println("end:" + time.Now().String())
}
