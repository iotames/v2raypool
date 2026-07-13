package webserver

import (
	"encoding/json"

	vp "github.com/iotames/v2raypool"
)

// TunnelStart 启动隧道代理池
func TunnelStart() []byte {
	result := BaseResult{}
	err := vp.StartTunnelPool()
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("隧道代理池已启动")
	return result.Bytes()
}

// TunnelStop 停止隧道代理池
func TunnelStop() []byte {
	result := BaseResult{}
	err := vp.StopTunnelPool()
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("隧道代理池已停止")
	return result.Bytes()
}

// TunnelStatus 获取隧道代理池运行状态
func TunnelStatus() []byte {
	tp := vp.GetTunnelPool()
	if tp == nil {
		dt := map[string]interface{}{
			"running":      false,
			"port":         0,
			"max_delay_ms": 0,
			"node_count":   0,
		}
		result := NewListData(dt, 0)
		by, _ := json.Marshal(result)
		return by
	}
	status := tp.GetStatus()
	result := NewListData(status, 1)
	by, _ := json.Marshal(result)
	return by
}
