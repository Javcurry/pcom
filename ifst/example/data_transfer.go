package main

import (
	"fmt"

	"hago-plat/pcom/ifst"
)

// SA ...
type SA struct {
	Data int64  `ifst:"data"`
	Str  string `ifst:"str2"`
}

func main() {
	data := map[string]interface{}{
		"data": 999,
		"str2": "hello",
	}

	var sa SA
	err := ifst.Transfer(data, &sa)
	if err != nil {
		fmt.Printf("transfer error: %v\n", err)
		return
	}

	fmt.Printf("transfer data: %+v", sa)
}
