package v2raypool

// 隧道代理池默认配置常量
const (
	DEFAULT_TUNNEL_PORT          = 1080
	DEFAULT_TUNNEL_ENABLE        = false
	DEFAULT_TUNNEL_MAX_DELAY_MS  = 230
	DEFAULT_TUNNEL_REFRESH_INTERVAL = 1200 // 隧道代理池测速刷新间隔（秒）
)

// TunnelConfig 隧道代理池配置
// 通过读取现有的节点订阅数据，筛选测速合格的节点作为出口节点池，
// 对外暴露单一 HTTP 代理端口，每次请求随机选择一个出口节点转发流量。
type TunnelConfig struct {
	// 隧道代理池是否启用。true 时启动隧道代理 HTTP 服务
	Enable bool
	// 隧道代理池监听的 HTTP 端口，例如 1080，外部爬虫通过 http://127.0.0.1:1080 使用
	Port int
	// 节点最大延迟阈值（毫秒）。只有测速延迟小于此值的节点才会被加入隧道代理池
	// 默认 230ms，可根据实际网络环境调整
	MaxDelayMs int
	// 节点测速刷新间隔（秒）。隧道池每隔此时间对运行中节点重新测速并更新可用节点列表
	// 默认 1200 秒（20 分钟），测速结果写回 ProxyPool.speedMap 供全局共享
	RefreshInterval int
}
