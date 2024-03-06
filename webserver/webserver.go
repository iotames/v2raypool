package webserver

import (
	"fmt"

	"github.com/iotames/glayui/web"
	"github.com/iotames/miniutils"
	"github.com/iotames/v2raypool/conf"
)

func RunWebServer() {
	addr := fmt.Sprintf(":%d", conf.GetConf().WebServerPort)
	fmt.Printf("-----启动Web服务器。监听地址(%s)------\n", addr)
	s := web.NewEasyServer(addr)
	setRouter(s)
	miniutils.StartBrowserByUrl(`http://127.0.0.1` + addr)
	s.ListenAndServe()
}
