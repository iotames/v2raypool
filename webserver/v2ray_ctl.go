package webserver

import (
	"encoding/json"
	"fmt"

	"github.com/iotames/miniutils"
	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
)

// CopyV2ray 复制一份v2ray的json配置文件。并更改入站端口号。
func CopyV2ray(oldConFile, newConFile, protocol string, localPort int, globalProxy bool) []byte {
	cf := conf.GetConf()
	result := BaseResult{}
	if !miniutils.IsPathExists(oldConFile) {
		result.Fail(fmt.Sprintf("原文件%s不存在", oldConFile), 400)
		return result.Bytes()
	}
	if miniutils.IsPathExists(newConFile) {
		result.Fail(fmt.Sprintf("目标文件%s已存在，请更换文件名。", newConFile), 400)
		return result.Bytes()
	}
	if localPort == 0 {
		result.Fail("本地入站端口不能为空", 400)
		return result.Bytes()
	}
	if miniutils.GetIndexOf(protocol, cf.GetOkInboundProtocols()) == -1 {
		result.Fail("入站协议不正确", 400)
		return result.Bytes()
	}
	if protocol == "socks5" {
		protocol = "socks"
	}
	err := vp.GetProxyPool().CheckLocalPort([]int{localPort})
	if err != nil {
		result.Fail(err.Error(), 400)
		return result.Bytes()
	}
	// 从原文件复制新文件
	oldConf := vp.NewJsonConfigFromFile(oldConFile)
	confv4 := vp.V2rayConfigV4{}
	err = oldConf.Decode(&confv4)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	confv4.Inbounds = []vp.V2rayInbound{vp.NewV2rayInboundV4(protocol, localPort)}
	if globalProxy {
		confv4.Routing = nil
		var newoutbds []vp.V2rayOutbound
		for _, outbdn := range confv4.Outbounds {
			if outbdn.Protocol == "freedom" || outbdn.Tag == "DIRECT" {
				continue
			}
			newoutbds = append(newoutbds, outbdn)
		}
		confv4.Outbounds = newoutbds
	}

	err = oldConf.SetContent(confv4)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	err = oldConf.SaveToFile(newConFile)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	return RunV2ray(newConFile, "复制启动成功")
}

func DeleteV2ray(pid int) []byte {
	result := BaseResult{}
	if pid == 0 {
		result.Fail("pid 不能为空", 400)
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

func RestartV2ray(pid int, configFile string) []byte {
	result := BaseResult{}
	pp := vp.GetProxyPool()
	err := pp.DeleteV2rayServer(pid)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	return RunV2ray(configFile, "重启成功")
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
	cf := conf.GetConf()
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

		vapiport := cf.V2rayApiPort
		sysport := cf.GetHttpProxyPort()

		for i, inbd := range vs.GetV2rayConfigV4().Inbounds {
			port := inbd.Port
			proto := inbd.Protocol
			if port == vapiport {
				proto = "grpc"
				runmode = "动态代理池"
			}
			addStr := fmt.Sprintf("%s:%d", proto, port)
			if port == sysport {
				runmode = "系统代理"
			}
			if i > 0 {
				addStr = ", " + addStr
			}
			localPorts += addStr
		}
		if runmode == "动态代理池" {
			localPorts += fmt.Sprintf(", %s:%s", cf.GetHttpProxyProtocol(), pp.GetLocalPortRange())
		}
		if runmode == "系统代理" {
			confile = vp.ROUTING_RULES_FILE + " -> " + confile
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
	Pid             int    `json:"pid"`
	InboundProtocol string `json:"inbound_protocol"`
	LocalPort       int    `json:"local_port"`
	OldConfigFile   string `json:"old_config_file"`
	ConfigFile      string `json:"config_file"`
	GlobalProxy     string `json:"global_proxy"`
}
