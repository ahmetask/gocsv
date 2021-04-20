package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ahmetask/gocsv"
	"reflect"
	"sync"
	"time"
)

type MyStruct struct {
	F1 *string
}

type Model struct {
	F1 []string
	F2 *int
	F3 *float64
	F4 *string
	F5 *MyStruct
}

func convertField(v string, k reflect.Kind, t reflect.Type) (interface{}, error) {
	return nil, errors.New(fmt.Sprintf("convertError data:%v kind:%v type:%v", v, k.String(), t.String()))
}

func main() {
	reader, err := gocsv.NewReader(gocsv.ReaderConfig{
		FilePath:        "./example/string/string.csv",
		ProducerBuffer:  100,
		Format:          Model{},
		ConvertFunction: convertField,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	readChannel := reader.Read()

	wg := &sync.WaitGroup{}
	wg.Add(10)
	fmt.Println("start:" + time.Now().String())

	for i := 0; i < 10; i++ {
		go func(w *sync.WaitGroup) {
			for r := range readChannel {
				if r.Exist() {
					if v, ok := r.Value().(*Model); ok {
						marshal, _ := json.Marshal(v)
						fmt.Println(string(marshal))
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
