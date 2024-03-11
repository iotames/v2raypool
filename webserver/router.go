package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/iotames/glayui/gtpl"
	"github.com/iotames/glayui/web"
	"github.com/iotames/miniutils"
	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
)

func setRouter(s *web.EasyServer) {
	tpl := gtpl.GetTpl()
	tpl.SetResourceDirPath("resource")

	s.AddHandler("GET", "/", func(ctx web.Context) {
		tpl.SetDataByTplFile("index.html", conf.GetConf(), ctx.Writer)
	})
	s.AddHandler("GET", "/api/nodes", func(ctx web.Context) {
		ctx.Writer.Write(GetNodes())
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
		dt := vp.ProxyNode{}
		err := getPostJson(ctx, &dt)
		if err != nil {
			return
		}
		ctx.Writer.Write(ActiveNode(dt.RemoteAddr))
	})

	s.AddHandler("POST", "/api/v2ray/run", func(ctx web.Context) {
		val, err := getPostJsonField(ctx, "config_file")
		if err != nil {
			return
		}
		configFile, ok := val.(string)
		if !ok {
			result := BaseResult{Msg: "config_file必须为字符串", Code: 500}
			ctx.Writer.Write(result.Bytes())
			return
		}
		ctx.Writer.Write(RunV2ray(configFile, "启动成功"))
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

func getPostJsonField(ctx web.Context, field string) (val any, err error) {
	dt := make(map[string]any)
	err = getPostJson(ctx, &dt)
	if err != nil {
		return
	}
	var ok bool
	val, ok = dt[field]
	if !ok {
		err = fmt.Errorf("post field %s not found", field)
		result := BaseResult{Msg: err.Error(), Code: 400}
		ctx.Writer.Write(result.Bytes())
	}
	return
}

func getPostJson(ctx web.Context, v any) error {
	postdata, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		result := BaseResult{Msg: err.Error(), Code: 500}
		ctx.Writer.Write(result.Bytes())
		return err
	}
	err = json.Unmarshal(postdata, v)
	if err != nil {
		result := BaseResult{Msg: err.Error(), Code: 500}
		ctx.Writer.Write(result.Bytes())
		return err
	}
	return err
}

func UpdateConf(dt map[string]string, fpath string) []byte {
	err := conf.UpdateConf(dt, fpath)
	result := BaseResult{}
	if err != nil {
		result.Fail("更新失败:"+err.Error(), 500)
		return result.Bytes()
	}
	// fmt.Printf("-----cf(%+v)---\n", dt)
	cf := conf.GetConf()
	cf.V2rayApiPort, err = strconv.Atoi(dt["VP_V2RAY_API_PORT"])
	if err != nil {
		result.Fail("VP_V2RAY_API_PORT 更新失败:"+err.Error(), 400)
		return result.Bytes()
	}
	cf.WebServerPort, err = strconv.Atoi(dt["VP_WEB_SERVER_PORT"])
	if err != nil {
		result.Fail("VP_WEB_SERVER_PORT 更新失败:"+err.Error(), 400)
		return result.Bytes()
	}
	cf.GrpcPort, err = strconv.Atoi(dt["VP_GRPC_PORT"])
	if err != nil {
		result.Fail("VP_GRPC_PORT 更新失败:"+err.Error(), 400)
		return result.Bytes()
	}
	cf.TestUrl = dt["VP_TEST_URL"]
	cf.SubscribeUrl = dt["VP_SUBSCRIBE_URL"]
	cf.SubscribeDataFile = dt["VP_SUBSCRIBE_DATA_FILE"]
	cf.V2rayPath = dt["VP_V2RAY_PATH"]
	cf.HttpProxy = dt["VP_HTTP_PROXY"]
	conf.SetConf(cf)
	result.Success("设置成功，重启应用后生效。")
	return result.Bytes()
}

func UnActiveNode(remoteAddr string) []byte {
	var err error
	pp := vp.GetProxyPool()
	ok := false
	for _, nd := range pp.GetNodes("") {
		if nd.RemoteAddr == remoteAddr {
			err = pp.UnActiveNode(nd)
			ok = true
			break
		}
	}
	result := BaseResult{}
	if !ok {
		result.Fail("找不到代理节点:"+remoteAddr, 400)
		return result.Bytes()
	}
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("禁用代理成功")
	return result.Bytes()
}

func ActiveNode(remoteAddr string) []byte {
	var err error
	pp := vp.GetProxyPool()
	ok := false
	for _, nd := range pp.GetNodes("") {
		if nd.RemoteAddr == remoteAddr {
			err = pp.ActiveNode(nd)
			ok = true
			break
		}
	}
	result := BaseResult{}
	if !ok {
		result.Fail("找不到代理节点:"+remoteAddr, 400)
		return result.Bytes()
	}
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("启用成功")
	return result.Bytes()
}

func StartNodes() []byte {
	result := BaseResult{}
	pp := vp.GetProxyPool()
	if pp.IsLock {
		result.Fail("系统繁忙，请稍候", 500)
		return result.Bytes()
	}
	err := pp.StartAll()
	if err != nil {
		result.Fail(err.Error(), 200)
		return result.Bytes()
	}
	result.Success("启动成功")
	return result.Bytes()
}

func TestNodes(testurl string) []byte {
	result := BaseResult{}
	pp := vp.GetProxyPool()
	if pp.IsLock {
		result.Fail("系统繁忙，请稍候", 500)
		return result.Bytes()
	}
	if len(vp.GetRunningNodes()) == 0 {
		msg := "没有可测速的代理节点。请先启动IP代理池"
		result.Fail(msg, 400)
		return result.Bytes()
	}
	oldConf := conf.GetConf()
	oldConf.TestUrl = testurl
	conf.SetConf(oldConf)
	pp.SetTestUrl(testurl)
	go pp.TestAll()
	result.Success("测速已开始，请稍候...")
	return result.Bytes()
}

func GetNodes() []byte {
	pp := vp.GetProxyPool()
	testDomain := miniutils.GetDomainByUrl(conf.GetConf().TestUrl)
	nds := pp.GetNodes(testDomain)
	if len(nds) == 0 {
		nds = pp.GetNodes("")
	}
	nds.SortBySpeed()
	activeNode := pp.GetActiveNode()
	var rows []map[string]any
	for _, n := range nds {
		isActive := false
		if activeNode.RemoteAddr == n.RemoteAddr {
			isActive = true
		}
		data := map[string]any{
			"index":       n.Index,
			"id":          n.Id,
			"protocol":    n.Protocol,
			"local_port":  n.LocalPort,
			"speed":       fmt.Sprintf("%.2f", n.Speed.Seconds()),
			"test_url":    n.TestUrl,
			"title":       n.Title,
			"local_addr":  pp.GetLocalAddr(n),
			"remote_addr": n.RemoteAddr,
			"is_running":  n.IsRunning(),
			"is_active":   isActive,
			"is_ok":       n.IsOk(),
			"test_at":     n.TestAt.Format("2006-01-02 15:04"),
		}
		// fmt.Printf("-----GetNodes---ndata(%+v)------\n", data)
		rows = append(rows, data)
	}
	result := NewListData(rows, len(rows))
	result.Code = 0
	b, err := json.Marshal(result)
	if err == nil {
		return b
	}
	res := BaseResult{
		Code: 500,
		Msg:  err.Error(),
	}
	return res.Bytes()
}
