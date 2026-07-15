package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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
		// 先渲染模板到缓冲区
		var tplBuf bytes.Buffer
		err := tpl.SetDataByTplFile("index.html", dt, &tplBuf)
		if err != nil {
			fmt.Printf("------err(%v)----\n", err)
		}
		// 服务端检查配置，有问题则写入可见横幅 + 弹窗脚本（纯 Go 字符串，无模板转义）
		raw := CheckConfig()
		var result struct {
			Data []ConfigCheckItem `json:"data"`
		}
		if e := json.Unmarshal(raw, &result); e == nil && len(result.Data) > 0 {
			var bHTML strings.Builder
			bHTML.WriteString(`<div id="startupAlert" style="margin:8px 16px;padding:14px 18px;border-radius:4px;background:#fff3f0;border:2px solid #ff4d4f;color:#cf1322;font-size:14px;line-height:1.8;font-weight:600">`)
			bHTML.WriteString(`⚠ 配置问题：<br>`)
			var alertLines []string
			for _, item := range result.Data {
				icon := "🟡"
				if item.Status == "missing" {
					icon = "🔴"
				} else if item.Status == "error" {
					icon = "🟠"
				}
				bHTML.WriteString(fmt.Sprintf(`%s <strong>%s</strong>：%s<br>`, icon, item.Label, item.Message))
				alertLines = append(alertLines, icon+" "+item.Label+"："+item.Message)
			}
			bHTML.WriteString(`<div style="margin-top:8px">请先配置必要参数，然后点击「▶ 启动」按钮，程序会自动初始化并启动节点。</div>`)
			bHTML.WriteString(`</div>`)
			// 横幅写在模板前面（页面最顶部）
			ctx.Writer.Write([]byte(bHTML.String()))
			// 内联 alert 脚本，JSON.Marshal 转义 JS 字符串，绝对安全
			alertLines = append(alertLines, "请先配置必要参数，然后点击「启动」按钮，程序会自动初始化并启动节点。")
			alertJSON, _ := json.Marshal(append([]string{"⚠ 配置检查"}, alertLines...))
			ctx.Writer.Write([]byte(fmt.Sprintf(`<script>(function(){var a=%s;if(a.length>1){window.alert(a.join("\n"));}})();</script>`, string(alertJSON))))
		}
		// 最后写模板内容
		ctx.Writer.Write(tplBuf.Bytes())
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

	s.AddHandler("POST", "/api/nodes/test-url", func(ctx web.Context) {
		dt := make(map[string]string)
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(SetTestUrlOnly(dt["TestUrl"]))
	})

	s.AddHandler("POST", "/api/nodes/start", func(ctx web.Context) {
		ctx.Writer.Write(StartNodes())
	})

	s.AddHandler("POST", "/api/nodes/subscribe", func(ctx web.Context) {
		req := RequestUpdateSubscribe{}
		err := getPostJson(ctx, &req)
		if err != nil {
			return
		}
		ctx.Writer.Write(UpdateSubscribe(req.SubscribeByProxy))
	})

	s.AddHandler("POST", "/api/node/active", func(ctx web.Context) {
		dt := RequestActiveNode{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(ActiveNode(dt.RemoteAddr, dt.GlobalProxy))
	})

	s.AddHandler("POST", "/api/node/delete", func(ctx web.Context) {
		dt := RequestNode{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(DeleteNode(dt.Index))
	})

	s.AddHandler("POST", "/api/v2ray/run", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(RunV2ray(dt.ConfigFile, "启动成功"))
	})
	s.AddHandler("POST", "/api/v2ray/copyrun", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		globalProxy := false
		if strings.EqualFold(dt.GlobalProxy, "on") {
			globalProxy = true
		}
		ctx.Writer.Write(CopyV2ray(dt.OldConfigFile, dt.ConfigFile, dt.InboundProtocol, dt.LocalPort, globalProxy))
	})
	s.AddHandler("POST", "/api/v2ray/restart", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(RestartV2ray(dt.Pid, dt.ConfigFile))
	})
	s.AddHandler("POST", "/api/v2ray/delete", func(ctx web.Context) {
		dt := V2rayServerData{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(DeleteV2ray(dt.Pid))
	})
	s.AddHandler("POST", "/api/node/unactive", func(ctx web.Context) {
		dt := vp.ProxyNode{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(UnActiveNode(dt.RemoteAddr))
	})

	// 代理池初始化 API
	s.AddHandler("GET", "/api/pool/init-status", func(ctx web.Context) {
		ctx.Writer.Write(PoolInitStatus())
	})
	s.AddHandler("POST", "/api/pool/init", func(ctx web.Context) {
		ctx.Writer.Write(PoolInit())
	})

	s.AddHandler("GET", "/api/setting/check", func(ctx web.Context) {
		ctx.Writer.Write(CheckConfig())
	})

	s.AddHandler("POST", "/api/setting/clearcache", func(ctx web.Context) {
		ctx.Writer.Write(ClearCache())
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

	s.AddHandler("POST", "/api/v2ray/routing-rules/update", func(ctx web.Context) {
		dt := RequestRoutingRules{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(UpdateV2rayRoutingRules(dt))
	})

	// 隧道代理池 API
	s.AddHandler("POST", "/api/tunnel/start", func(ctx web.Context) {
		ctx.Writer.Write(TunnelStart())
	})
	s.AddHandler("POST", "/api/tunnel/stop", func(ctx web.Context) {
		ctx.Writer.Write(TunnelStop())
	})
	s.AddHandler("GET", "/api/tunnel/status", func(ctx web.Context) {
		ctx.Writer.Write(TunnelStatus())
	})

	// 系统代理切换 API
	s.AddHandler("GET", "/api/sysproxy/status", func(ctx web.Context) {
		ctx.Writer.Write(SysProxyStatus())
	})
	s.AddHandler("GET", "/api/sysproxy/check", func(ctx web.Context) {
		ctx.Writer.Write(SysProxyCheck())
	})
	s.AddHandler("POST", "/api/sysproxy/switch", func(ctx web.Context) {
		dt := make(map[string]int)
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		reqBody, _ := json.Marshal(dt)
		ctx.Writer.Write(SysProxySwitch(reqBody))
	})
}

type HomePageData struct {
	conf.Conf
	TestedDomainList []string
	ConfigIssueLines []string
}

type RequestNode struct {
	Index int
}
type RequestActiveNode struct {
	RemoteAddr  string `json:"remote_addr"`
	GlobalProxy bool   `json:"global_proxy"`
}

type RequestUpdateSubscribe struct {
	SubscribeUrl     string `json:"subscribe_url"`
	SubscribeByProxy string `json:"subscribe_by_proxy"`
}
