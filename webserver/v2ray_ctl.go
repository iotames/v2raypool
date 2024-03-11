package webserver

import (
	"encoding/json"
	"fmt"

	"github.com/iotames/miniutils"
	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
)

func DeleteV2ray(dt V2rayServerData) []byte {
	result := BaseResult{}
	pid := dt.Pid
	if pid == 0 {
		result.Success("pid 不能为空")
		return result.Bytes()
	}
	err := vp.GetProxyPool().DeleteV2rayServer(pid)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("删除成功")
	return result.Bytes()
}

func RestartV2ray(dt V2rayServerData) []byte {
	result := BaseResult{}
	pp := vp.GetProxyPool()
	err := pp.DeleteV2rayServer(dt.Pid)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	return RunV2ray(dt.ConfigFile, "重启成功")
}

func RunV2ray(fpath, msg string) []byte {
	var err error
	result := BaseResult{}
	if !miniutils.IsPathExists(fpath) {
		result.Fail(fmt.Sprintf("找不到配置文件: %s", fpath), 400)
		return result.Bytes()
	}
	pp := vp.GetProxyPool()
	vconf := vp.NewJsonConfigFromFile(fpath)
	err = pp.CheckV2rayConfig(*vconf)
	if err != nil {
		result.Fail(err.Error(), 400)
		return result.Bytes()
	}
	vs := vp.NewV2ray(conf.GetConf().V2rayPath)
	err = vs.Start(fpath)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	err = pp.AddV2rayServer(vs)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success(msg)
	return result.Bytes()
}

func GetV2rayList() []byte {
	pp := vp.GetProxyPool()
	slist := pp.GetV2rayServerList()
	var rows []map[string]any
	for _, vs := range slist {
		pid := 0
		if vs.GetExeCmd() != nil {
			pid = vs.GetExeCmd().Process.Pid
		}
		jconf := vs.GetJsonConfig()
		confile := jconf.GetFilepath()
		localPorts := ""
		runmode := "个性配置"
		vports := vs.GetLocalPortList()
		for i, port := range vports {
			addStr := fmt.Sprintf("%d", port)
			if i > 0 {
				addStr = "," + addStr
			}
			localPorts += addStr
		}
		if confile == "" {
			runmode = "动态代理池"
			localPorts += "," + pp.GetLocalPortRange()
		} else {
			if vs.GetLocalPort() > 0 {
				runmode = "系统代理"
				confile = vp.ROUTING_RULES_FILE + " -> " + confile
			}
		}
		data := map[string]any{
			"pid":         pid,
			"run_mode":    runmode,
			"local_ports": localPorts,
			"config_file": confile,
			"config_json": jconf.String(),
		}
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

type V2rayServerData struct {
	Pid        int    `json:"pid"`
	ConfigFile string `json:"config_file"`
}
