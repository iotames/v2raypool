package webserver

import (
	// "github.com/iotames/glayui/component"
	"encoding/json"
	"fmt"

	"github.com/iotames/glayui/gtpl"
	"github.com/iotames/glayui/web"
	vp "github.com/iotames/v2raypool"
)

func setRouter(s *web.EasyServer) {
	tpl := gtpl.GetTpl()
	tpl.SetResourceDirPath("resource")

	s.AddHandler("GET", "/", func(ctx web.Context) {
		tpl.SetDataByTplFile("index.html", nil, ctx.Writer)
	})
	s.AddHandler("GET", "/api/nodes", func(ctx web.Context) {
		ctx.Writer.Write(GetNodes())
	})
	s.AddHandler("POST", "/api/nodes/test", func(ctx web.Context) {
		ctx.Writer.Write(TestNodes())
	})
	s.AddHandler("POST", "/api/nodes/start", func(ctx web.Context) {
		ctx.Writer.Write(StartNodes())
	})
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
func TestNodes() []byte {
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
	go pp.TestAll()
	result.Success("测速已开始，请稍候...")
	return result.Bytes()
}

func GetNodes() []byte {
	pp := vp.GetProxyPool()
	nds := pp.GetNodes("")
	nds.SortBySpeed()
	var rows []map[string]any
	for i, n := range nds {
		isActive := false
		if i == 6 {
			isActive = true
		}
		runState := "已停止"
		if n.IsRunning() {
			runState = "运行中"
		}
		data := map[string]any{
			"index":       n.Index,
			"id":          n.Id,
			"local_port":  n.LocalPort,
			"speed":       n.Speed.Seconds(),
			"title":       n.Title,
			"local_addr":  pp.GetLocalAddr(n),
			"remote_addr": n.RemoteAddr,
			"is_running":  runState,
			"is_active":   isActive,
			"is_ok":       n.IsOk(),
			"test_at":     n.TestAt.Format("2006-01-02 15:04"),
		}
		fmt.Printf("-----GetNodes---ndata(%+v)------\n", data)
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
