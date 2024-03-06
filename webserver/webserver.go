package webserver

import (
	"fmt"

	"github.com/iotames/glayui/web"
	"github.com/iotames/miniutils"
)

func RunWebServer(webPort int) {
	addr := fmt.Sprintf(":%d", webPort)
	fmt.Printf("-----启动Web服务器。监听地址(%s)------\n", addr)
	s := web.NewEasyServer(addr)
	setRouter(s)
	err := miniutils.StartBrowserByUrl(`http://127.0.0.1` + addr)
	if err != nil {
		fmt.Println("StartBrowserByUrl error: " + err.Error())
	}
	s.ListenAndServe()
}
