package webserver

import (
	"fmt"
	"net"
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
	result := BaseResult{}
	cf := conf.GetConf()
	var err error
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
	// 校验：当订阅地址是本地文件路径（非 http/https 网络地址）时，
	// 订阅地址(VP_SUBSCRIBE_URL) 和 订阅数据文件(VP_SUBSCRIBE_DATA_FILE) 的值必须一致
	if _, ok := dt["VP_SUBSCRIBE_URL"]; ok {
		subUrl := cf.SubscribeUrl
		subFile := cf.SubscribeDataFile
		if subUrl != "" && !strings.HasPrefix(subUrl, "http://") && !strings.HasPrefix(subUrl, "https://") {
			if subFile == "" {
				result.Fail("订阅地址(VP_SUBSCRIBE_URL)是文件路径时，订阅数据文件(VP_SUBSCRIBE_DATA_FILE)不能为空", 400)
				return result.Bytes()
			}
			if subFile != subUrl {
				result.Fail("订阅地址(VP_SUBSCRIBE_URL)是文件路径时，必须与订阅数据文件(VP_SUBSCRIBE_DATA_FILE)的值一致", 400)
				return result.Bytes()
			}
		}
	}
	if v, ok := dt["VP_V2RAY_PATH"]; ok && v != "" {
		cf.V2rayPath = v
	}
	if v, ok := dt["VP_HTTP_PROXY"]; ok && v != "" {
		cf.HttpProxy = v
	}
	if v, ok := dt["VP_TUNNEL_MAX_DELAY"]; ok {
		if n, err2 := strconv.Atoi(v); err2 == nil && n > 0 {
			if n < 50 {
				result.Fail("VP_TUNNEL_MAX_DELAY 更新失败: 延迟阈值不可小于50 ms", 400)
				return result.Bytes()
			}
			cf.TunnelMaxDelay = n
			if tp := vp.GetTunnelPool(); tp != nil && tp.IsRunning() {
				tp.SetMaxDelay(n)
			}
		}
	}
	if v, ok := dt["VP_TUNNEL_REFRESH_INTERVAL"]; ok {
		if n, err2 := strconv.Atoi(v); err2 == nil {
			if n < conf.MIN_REFRESH_INTERVAL {
				result.Fail(fmt.Sprintf("VP_TUNNEL_REFRESH_INTERVAL 更新失败: 测速间隔不可小于%d s", conf.MIN_REFRESH_INTERVAL), 400)
				return result.Bytes()
			}
			cf.TunnelRefreshInterval = n
			if tp := vp.GetTunnelPool(); tp != nil && tp.IsRunning() {
				tp.SetRefreshInterval(n)
			}
		}
	}
	if v, ok := dt["VP_TUNNEL_PORT"]; ok {
		if n, err2 := strconv.Atoi(v); err2 == nil && n > 0 && n <= 65535 {
			// 端口变更需要重启隧道池生效
			if tp := vp.GetTunnelPool(); tp != nil && tp.IsRunning() {
				// 先探测新端口是否可用，避免 Stop 后无法重启
				probeLn, probeErr := net.Listen("tcp", fmt.Sprintf(":%d", n))
				if probeErr != nil {
					result.Fail(fmt.Sprintf("VP_TUNNEL_PORT 更新失败: 端口 %d 被占用", n), 500)
					return result.Bytes()
				}
				probeLn.Close()
				cf.TunnelPort = n
				tp.SetPort(n)
				tp.Stop()
				tp.Start()
			} else {
				cf.TunnelPort = n
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
	// 延迟阈值变更后，同步刷新隧道池节点列表，确保前端的节点数实时联动
	if _, ok := dt["VP_TUNNEL_MAX_DELAY"]; ok {
		if tp := vp.GetTunnelPool(); tp != nil && tp.IsRunning() {
			tp.RefreshNodes()
		}
	}
	// 所有校验通过后，写入 .env 文件
	if err := conf.UpdateConf(dt, fpath); err != nil {
		result.Fail("写入文件失败:"+err.Error(), 500)
		return result.Bytes()
	}
	result.Success("设置成功，已即时生效。")
	return result.Bytes()
}

// ConfigCheckItem 配置检查项
type ConfigCheckItem struct {
	Field   string `json:"field"`
	Label   string `json:"label"`
	Status  string `json:"status"` // missing / error / warning
	Message string `json:"message"`
}

// CheckConfig 检查当前配置，返回所有问题项
func CheckConfig() []byte {
	cf := conf.GetConf()
	var items []ConfigCheckItem

	// 检查 v2ray 路径
	if cf.V2rayPath == "" {
		items = append(items, ConfigCheckItem{
			Field: "VP_V2RAY_PATH", Label: "v2ray核心文件",
			Status: "missing", Message: "v2ray核心文件路径未配置",
		})
	} else if !miniutils.IsPathExists(cf.V2rayPath) {
		items = append(items, ConfigCheckItem{
			Field: "VP_V2RAY_PATH", Label: "v2ray核心文件",
			Status: "missing", Message: "v2ray核心文件不存在，请下载v2ray核心文件",
		})
	}

	// 检查订阅地址
	if cf.SubscribeUrl == "" {
		items = append(items, ConfigCheckItem{
			Field: "VP_SUBSCRIBE_URL", Label: "订阅地址",
			Status: "warning", Message: "订阅地址未配置，无法自动获取代理节点（可手动添加v2ray服务）",
		})
	}

	// 检查测速地址
	if cf.TestUrl == "" {
		items = append(items, ConfigCheckItem{
			Field: "VP_TEST_URL", Label: "测速地址",
			Status: "missing", Message: "测速地址未配置",
		})
	}

	// 检查 HttpProxy 格式
	if cf.HttpProxy != "" {
		protcls := cf.GetOkInboundProtocols()
		spt := strings.Split(cf.HttpProxy, ":")
		if len(spt) != 3 {
			items = append(items, ConfigCheckItem{
				Field: "VP_HTTP_PROXY", Label: "系统代理地址",
				Status: "error", Message: "系统代理地址格式不正确，应为 protocol:host:port 格式",
			})
		} else if miniutils.GetIndexOf(spt[0], protcls) == -1 {
			items = append(items, ConfigCheckItem{
				Field: "VP_HTTP_PROXY", Label: "系统代理地址",
				Status: "error", Message: "系统代理协议仅支持 http/socks/socks5",
			})
		}
	}

	// 检查运行时目录
	if cf.RuntimeDir == "" {
		items = append(items, ConfigCheckItem{
			Field: "VP_RUNTIME_DIR", Label: "运行时目录",
			Status: "error", Message: "运行时目录未配置",
		})
	} else if !miniutils.IsPathExists(cf.RuntimeDir) {
		items = append(items, ConfigCheckItem{
			Field: "VP_RUNTIME_DIR", Label: "运行时目录",
			Status: "warning", Message: "运行时目录不存在，启动时将自动创建",
		})
	}

	result := NewListData(items, len(items))
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
