package webserver

import (
	"fmt"
	// "github.com/iotames/glayui/component"
	"github.com/iotames/glayui/gtpl"
	"github.com/iotames/glayui/web"
	"github.com/iotames/v2raypool/conf"
)

func RunWebServer() {
	addr := fmt.Sprintf(":%d", conf.GetConf().WebServerPort)
	fmt.Printf("-----启动Web服务器。监听地址(%s)------\n", addr)
	tpl := gtpl.GetTpl()
	tpl.SetResourceDirPath("resource")
	s := web.NewEasyServer(addr)
	s.AddHandler("GET", "/", func(ctx web.Context) {
		tpl.SetDataByTplFile("index.html", nil, ctx.Writer)
	})
	s.ListenAndServe()
}
