package main

import (
	vp "github.com/iotames/v2raypool"
)

func main() {
	var err error
	// https://hostloc.com/thread-649363-1-1.html
	c := vp.NewV2rayApiClientV5("127.0.0.1:5059")
	if c.Dial() == nil {
		defer c.Close()
	}

	// err = c.AddInbound(10078, "T10078")
	// if err != nil {
	// 	panic(err)
	// }

	err = c.RemoveInbound("T10078")
	if err != nil {
		panic(err)
	}

}
