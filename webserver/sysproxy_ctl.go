package webserver

import (
	"encoding/json"

	vp "github.com/iotames/v2raypool"
)

// SysProxySwitch 切换系统代理模式
// req: {"type": 0|1|2, "node_idx": 123, "global": true}
// type: 0=无代理, 1=固定节点, 2=隧道代理
// global: 单节点模式下是否全局代理（true=全局/false=智能分流），仅 type=1 有效，默认 true
func SysProxySwitch(reqBody []byte) []byte {
	result := BaseResult{}
	var req struct {
		Type    int  `json:"type"`
		NodeIdx int  `json:"node_idx"`
		Global  *bool `json:"global"` // nil = 默认全局
	}
	if err := json.Unmarshal(reqBody, &req); err != nil {
		result.Fail("请求参数格式错误: "+err.Error(), 400)
		return result.Bytes()
	}
	if req.Type < 0 || req.Type > 2 {
		result.Fail("代理类型无效，有效值: 0=无代理, 1=固定节点, 2=隧道代理", 400)
		return result.Bytes()
	}
	isGlobal := true
	if req.Global != nil {
		isGlobal = *req.Global
	}
	if err := vp.SetSysProxy(vp.SysProxyType(req.Type), req.NodeIdx, isGlobal); err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("系统代理切换成功")
	return result.Bytes()
}

// SysProxyStatus 获取当前系统代理状态
func SysProxyStatus() []byte {
	status := vp.GetSysProxyStatus()
	result := NewListData(status, 1)
	by, _ := json.Marshal(result)
	return by
}

// SysProxyCheck 系统代理切换前预检
// 返回当前环境是否满足切换条件
func SysProxyCheck() []byte {
	type CheckResult struct {
		HasRunningNodes bool `json:"has_running_nodes"`
		TunnelRunning   bool `json:"tunnel_running"`
		TunnelAvailable bool `json:"tunnel_available"`
		TunnelNodeCount int  `json:"tunnel_node_count"`
		ActiveNodeIdx   int  `json:"active_node_idx"`
	}
	check := CheckResult{}
	pp := vp.GetProxyPool()
	nds := pp.GetNodes("")
	for _, nd := range nds {
		if nd.IsRunning() {
			check.HasRunningNodes = true
			if check.ActiveNodeIdx < 0 {
				check.ActiveNodeIdx = nd.Index
			}
		}
	}
	tp := vp.GetTunnelPool()
	if tp != nil {
		status := tp.GetStatus()
		if v, ok := status["running"].(bool); ok {
			check.TunnelRunning = v
		}
		if v, ok := status["node_count"].(int); ok {
			check.TunnelNodeCount = v
		}
		check.TunnelAvailable = check.TunnelRunning && check.TunnelNodeCount > 0
	}
	result := NewListData(check, 1)
	by, _ := json.Marshal(result)
	return by
}
