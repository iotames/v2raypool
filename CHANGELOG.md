## 升级日志

### v1.9.3

- Makefile 支持全平台交叉编译（Linux/macOS/Windows），简化 CI 构建流程
- 升级 easyconf v1.2.2 → v1.2.3，提升配置读写兼容性
- **配置缺失不阻塞启动**：v2ray 路径/订阅地址缺失不再 panic，允许通过 WebUI 配置后再使用
- 新增 `GET /api/setting/check` 检查配置问题，WebUI 首页红色横幅 + 弹窗提示缺失项
- 新增 `GET /api/pool/init-status` + `POST /api/pool/init` 手动初始化代理池
- 设置弹窗中缺失字段红框高亮，前端即时反馈
- 替换 `godotenv` 为 `easyconf.UpdateByMap` 写入 .env，提升一致性
- **系统代理状态追踪重构**：新增 `sysProxyGlobal` 变量，状态栏显示4种状态（关闭/单节点全局/单节点智能分流/隧道全局），系统代理切换 API 支持 `global` 参数
- 修复 `DeleteV2rayServer` 未清理 `p.activeCmd` 导致后续 `killActiveNode` 报 `Access is denied`
- 修复 `AddNode` 端口冲突时 `panic` 导致前端无法收到错误提示
- 修复 `SubscribeUrl` 为空时未检查 `SubscribeDataFile` 兜底逻辑
- 修复 Windows 上已退出进程 `TerminateProcess` 返回 `Access is denied` 导致无法切换代理模式
- v2ray 列表 UI 按钮按角色（个性配置/系统代理/动态代理池）分粒度控制
- **修复 `/api/v2ray/list` 出现重复系统代理条目**：提取 `cleanServerMapSysPort()` 方法，在 `Delete()`、`ActiveNode()`、`KillAllNodes()` 中统一清理 `serverMap` 中残留的系统代理进程记录
- README 更新隧道池架构图，配置说明明确订阅方式二选一

### v1.9.2

- 前端重构：移除 layui 依赖，纯原生 JS/CSS/HTML 重写，加载更快、体积更小
- WebUI 卡片化重排：隧道状态、节点信息卡片式展示，关键数值点击即可修改
- 隧道池端口热更新：修改 VP_TUNNEL_PORT 后即时生效，无需重启服务
- 隧道池配置即时热更新：延迟阈值、测速间隔等配置修改后同步至运行中实例
- 弹窗遮罩层关闭优化：点击外部关闭改为 mousedown 事件，避免选择文字时误关闭
- 配置修改安全增强：所有校验移至赋值前执行，校验失败不污染 .env 文件
- 端口变更安全增强：Stop 前先探测新端口可用性，被占用时拒绝变更并报错
- 节点数实时联动：修改延迟阈值后隧道池立即重新筛选节点，前端节点数即时更新
- RefreshInterval 启动约束：NewTunnelPool 启动时强制 RefreshInterval >= 300s，与运行时行为统一
- 新增 MIN_REFRESH_INTERVAL 常量（300s），统一测速间隔最小值的校验入口

### v1.9.1

- 支持配置动态更新：WebUI 在线修改隧道阈值、测速地址等配置，实时写入文件并同步至运行中实例
- 隧道代理池测速优化：刷新时跳过 24h 内有效测速结果，减少重复请求
- 修复系统代理状态一致性：Delete/KillAllNodes/UnActiveNode 时自动清零代理状态
- gRPC SetTestUrl 持久化到配置文件
- conf_ctl.go 改为部分字段可选更新，不再强制全量覆盖
- WebUI 优化：下拉选择域名自动切换测速地址，隧道延迟阈值可点击直接修改

### v1.9.2

- 支持Clash的.yml配置文件。
- 支持保存节点测速数据到本地文件，方便下次重启时读取。

### v1.7.3

- 添加Windows应用图标和软件信息

### v1.7.0

- 更新配置组件

### v1.6.4

1. 新增删除节点功能
2. 更新ss://节点格式解析规则

### v1.6.3

- 添加清除应用缓存功能

### v1.6.2

- 修复读取测速结果导致全局代理不可用的BUG

### v1.6.1

1. 节点测速结果持久保存到文件中，以便下次读取
2. 允许使用自定义HTTP代理更新订阅

### v1.6.0

- 支持 trojan 协议

### v1.5.0

1. 支持shadowsocks(ss://)协议
2. 完善vmess节点解析规则
