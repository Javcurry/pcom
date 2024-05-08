package main

import (
	"fmt"
	"hago-plat/pcom/ifst"
)

// STA ...
type STA struct {
	Int  int64
	Bool bool
	Map  map[string]float64
	MP   map[string]*STA
}

// STB ...
type STB struct {
	Int  int64
	Bool bool
	Map  map[string]float64
	MP   map[string]*STB
}

func main() {
	a := &STA{
		Int:  999,
		Bool: true,
		Map: map[string]float64{
			"a": 0.5,
			"b": 1.5,
		},
		MP: map[string]*STA{
			"haha": {Int: 555},
		},
	}
	b := &STB{}

	err := ifst.Transfer(a, b)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println(a)
	fmt.Println(b)
}
