package webserver

import (
	"fmt"

	"github.com/iotames/glayui/web"
)

func NewWebServer(webPort int) *web.EasyServer {
	addr := fmt.Sprintf(":%d", webPort)
	fmt.Printf("-----启动Web服务器。监听地址(%s)------\n", addr)
	s := web.NewEasyServer(addr)
	setRouter(s)
	return s
}
