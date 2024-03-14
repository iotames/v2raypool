package webserver

import (
	"encoding/json"
	"fmt"

	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
)

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

// ActiveNode 启用一个远程节点作为系统代理
func ActiveNode(remoteAddr string, globalProxy bool) []byte {
	var err error
	pp := vp.GetProxyPool()
	ok := false
	for _, nd := range pp.GetNodes("") {
		if nd.RemoteAddr == remoteAddr {
			err = pp.ActiveNode(nd, globalProxy)
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

func GetNodes(domain string) []byte {
	pp := vp.GetProxyPool()
	// if domain == "" {
	// 	domain = miniutils.GetDomainByUrl(conf.GetConf().TestUrl)
	// }
	nds := pp.GetNodes(domain)
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
