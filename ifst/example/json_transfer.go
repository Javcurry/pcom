package main

import (
	"encoding/json"
	"fmt"

	"hago-plat/pcom/ifst"
)

// KindGetter ...
type KindGetter struct {
	Kind string `json:"kind"`
}

// Car ...
type Car struct {
	Kind   string `json:"kind"`
	Doors  int64  `json:"doors"`
	Wheels int64  `json:"wheels`
}

func main() {
	jsonData := `
{
	"kind": "Car",
	"doors": 4,
	"wheels": 4
}`

	var data interface{}
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		fmt.Printf("json unmarshal fail: %v", err)
		return
	}

	tf := ifst.NewTransformer()
	tf.TagName = "json"

	var kindGetter KindGetter
	err = tf.Transfer(data, &kindGetter)
	if err != nil {
		fmt.Printf("transfer to KindGetter fail: %v", err)
		return
	}

	/*
		switch kindGetter.Kind {
			...
		}
	*/

	var car Car
	err = tf.Transfer(data, &car)
	if err != nil {
		fmt.Printf("transfer to Car fail: %v", err)
		return
	}
}
