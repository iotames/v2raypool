# TODO

## 目标
启动后缺少必要配置时，WebUI 自动弹出警告对话框 + 顶部红色横幅。

## 已完成
- 后端：配置缺失不再 panic，新增 `CheckConfig()` API、手动初始化 `InitProxyPool()`
- 前端：`index.html` 保留 `configAlertBanner` 容器，`app.js` 调用 API 检测配置问题并弹窗+横幅
- `/` 路由处理器直出 alert 脚本作为兜底

## 待验证
- 浏览器打开 WebUI 是否弹出 alert（`cd main && go build -o v2raypool.exe .` → 运行 → 访问 `http://127.0.0.1:8087`）
- 若无效：F12 查看 Console JS 错误

## 已知问题
- `glayui/gtpl` 模板引擎可能 HTML-转义输出导致 JS 语法错误
- 无法通过 curl 验证 alert 弹窗（需浏览器环境）
