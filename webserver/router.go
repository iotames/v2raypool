package webserver

import (
	"fmt"

	"github.com/iotames/glayui/web"

	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
)

func setRouter(s *web.EasyServer) {
	tpl := GetTpl()
	s.AddHandler("GET", "/", func(ctx web.Context) {
		dt := HomePageData{Conf: conf.GetConf()}
		pp := vp.GetProxyPool()
		dt.TestedDomainList = pp.GetTestedDomainList()
		err := tpl.SetDataByTplFile("index.html", dt, ctx.Writer)
		if err != nil {
			fmt.Printf("------err(%v)----\n", err)
		}
	})
	s.AddHandler("GET", "/api/nodes", func(ctx web.Context) {
		domain := ctx.Request.URL.Query().Get("domain")
		ctx.Writer.Write(GetNodes(domain))
	})
	s.AddHandler("GET", "/api/v2ray/list", func(ctx web.Context) {
		ctx.Writer.Write(GetV2rayList())
	})
	s.AddHandler("POST", "/api/nodes/test", func(ctx web.Context) {
		dt := make(map[string]string)
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(TestNodes(dt["TestUrl"]))
	})
	s.AddHandler("POST", "/api/nodes/start", func(ctx web.Context) {
		ctx.Writer.Write(StartNodes())
	})
	s.AddHandler("POST", "/api/node/active", func(ctx web.Context) {
		dt := RequestActiveNode{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(ActiveNode(dt.RemoteAddr, dt.GlobalProxy))
	})

	s.AddHandler("POST", "/api/v2ray/run", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(RunV2ray(dt.ConfigFile, "启动成功"))
	})

	s.AddHandler("POST", "/api/v2ray/restart", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(RestartV2ray(dt))
	})
	s.AddHandler("POST", "/api/v2ray/delete", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(DeleteV2ray(dt))
	})
	s.AddHandler("POST", "/api/node/unactive", func(ctx web.Context) {
		dt := vp.ProxyNode{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(UnActiveNode(dt.RemoteAddr))
	})

	s.AddHandler("POST", "/api/setting/update", func(ctx web.Context) {
		// envfile := ctx.Server.GetData("ENV_FILE").Value.(string)
		// fmt.Println(envfile)
		dt := make(map[string]string)
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(UpdateConf(dt, conf.GetConf().EnvFile))
	})
}

type HomePageData struct {
	conf.Conf
	TestedDomainList []string
}

type RequestActiveNode struct {
	RemoteAddr  string `json:"remote_addr"`
	GlobalProxy bool   `json:"global_proxy"`
}
