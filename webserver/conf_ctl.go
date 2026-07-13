package webserver

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iotames/miniutils"
	vp "github.com/iotames/v2raypool"
	"github.com/iotames/v2raypool/conf"
)

// UpdateV2rayRoutingRules 更改配置文件的路由规则
// 更改 .env文件。删除 routing.rules.json 文件。
func UpdateV2rayRoutingRules(dt RequestRoutingRules) []byte {
	result := BaseResult{}

	// 更改 .env 文件
	updatedt := make(map[string]string, 4)
	updatedt["VP_DIRECT_DOMAIN_LIST"] = strings.Join(dt.DirectDomainList, ",")
	updatedt["VP_DIRECT_IP_LIST"] = strings.Join(dt.DirectIpList, ",")
	updatedt["VP_PROXY_DOMAIN_LIST"] = strings.Join(dt.ProxyDomainList, ",")
	updatedt["VP_PROXY_IP_LIST"] = strings.Join(dt.ProxyIpList, ",")
	cf := conf.GetConf()
	err := conf.UpdateConf(updatedt, cf.EnvFile)
	if err != nil {
		result.Fail("更新失败:"+err.Error(), 500)
		return result.Bytes()
	}
	// 更新配置
	cf.DirectDomainList = dt.DirectDomainList
	cf.DirectIpList = dt.DirectIpList
	cf.ProxyDomainList = dt.ProxyDomainList
	cf.ProxyIpList = dt.ProxyIpList
	conf.SetConf(cf)
	// 如存在 routing.rules.json 文件则删除。
	if miniutils.IsPathExists(vp.ROUTING_RULES_FILE) {
		err = os.Remove(vp.ROUTING_RULES_FILE)
		if err != nil {
			result.Fail("删除routing.rules.json文件失败:"+err.Error(), 500)
			return result.Bytes()
		}
	}

	result.Success("路由规则更新成功.请重新启用代理节点。")
	return result.Bytes()
}

// UpdateConf 更改.env配置文件
func UpdateConf(dt map[string]string, fpath string) []byte {
	err := conf.UpdateConf(dt, fpath)
	result := BaseResult{}
	if err != nil {
		result.Fail("更新失败:"+err.Error(), 500)
		return result.Bytes()
	}
	// fmt.Printf("-----cf(%+v)---\n", dt)
	cf := conf.GetConf()
	if v, ok := dt["VP_V2RAY_API_PORT"]; ok && v != "" {
		cf.V2rayApiPort, err = strconv.Atoi(v)
		if err != nil {
			result.Fail("VP_V2RAY_API_PORT 更新失败:"+err.Error(), 400)
			return result.Bytes()
		}
	}
	if v, ok := dt["VP_WEB_SERVER_PORT"]; ok && v != "" {
		cf.WebServerPort, err = strconv.Atoi(v)
		if err != nil {
			result.Fail("VP_WEB_SERVER_PORT 更新失败:"+err.Error(), 400)
			return result.Bytes()
		}
	}
	if v, ok := dt["VP_GRPC_PORT"]; ok && v != "" {
		cf.GrpcPort, err = strconv.Atoi(v)
		if err != nil {
			result.Fail("VP_GRPC_PORT 更新失败:"+err.Error(), 400)
			return result.Bytes()
		}
	}
	if v, ok := dt["VP_TEST_URL"]; ok && v != "" {
		cf.TestUrl = v
	}
	if v, ok := dt["VP_SUBSCRIBE_URL"]; ok && v != "" {
		cf.SubscribeUrl = v
	}
	if v, ok := dt["VP_SUBSCRIBE_DATA_FILE"]; ok && v != "" {
		cf.SubscribeDataFile = v
	}
	if v, ok := dt["VP_V2RAY_PATH"]; ok && v != "" {
		cf.V2rayPath = v
	}
	if v, ok := dt["VP_HTTP_PROXY"]; ok && v != "" {
		cf.HttpProxy = v
	}
	if v, ok := dt["VP_TUNNEL_MAX_DELAY"]; ok {
		if n, err2 := strconv.Atoi(v); err2 == nil && n > 0 {
			cf.TunnelMaxDelay = n
			// 同步更新运行中的隧道池
			if tp := vp.GetTunnelPool(); tp != nil && tp.IsRunning() {
				tp.SetMaxDelay(n)
			}
		}
	}
	conf.SetConf(cf)
	pp := vp.GetProxyPool()
	if _, ok := dt["VP_TEST_URL"]; ok {
		pp.SetTestUrl(cf.TestUrl)
	}
	if _, ok := dt["VP_SUBSCRIBE_URL"]; ok {
		pp.SetSubscribeUrl(cf.SubscribeUrl)
	}
	result.Success("设置成功，重启应用后生效。")
	return result.Bytes()
}

func ClearCache() []byte {
	result := BaseResult{}
	runtimedir := conf.GetConf().RuntimeDir
	flist, err := os.ReadDir(runtimedir)
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	for _, d := range flist {
		fullpath := filepath.Join(runtimedir, d.Name())
		if d.IsDir() {
			err = os.RemoveAll(fullpath)
		} else {
			err = os.Remove(fullpath)
		}
		if err != nil {
			// 忽略因无法删除打开中的日志文件而产生的报错
			if strings.Contains(err.Error(), ".log: The process cannot access the file because it is being used by another process") {
				err = nil
			}
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		result.Fail(err.Error(), 500)
		return result.Bytes()
	}
	result.Success("清除成功")
	return result.Bytes()
}

type RequestRoutingRules struct {
	DirectDomainList []string `json:"direct_domain_list"`
	DirectIpList     []string `json:"direct_ip_list"`
	ProxyDomainList  []string `json:"proxy_domain_list"`
	ProxyIpList      []string `json:"proxy_ip_list"`
}
