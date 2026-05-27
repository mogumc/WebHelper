<p align="center">
  <img src="./frontend/public/appicon.png" alt="WebHelper" width="300"/>
</p>

# WebHelper

基于 Wails v2 + Vue 3 + Element Plus 的网页调试助手桌面应用。

## 功能

- **网页调试** — 发送 HTTP 请求，支持所有方法、请求头、Cookie、代理、文件上传、请求日志
- **JS 调试** — 内置 JavaScript 执行引擎，支持 console.log 捕获和返回值
- **JSON 解析** — 树形/表格/原始三种视图，支持从网页调试一键发送
- **Socket 调试** — WebSocket/TCP 连接、收发消息、实时日志
- **多语言** — 内置简体中文、English、日本語、한국어，运行时切换

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go + Wails v2 + goja (JS引擎) + gorilla/websocket |
| 前端 | Vue 3 + Element Plus + Vite |
| 国际化 | 自定义 i18n，Go 后端加载，`go:embed` 打包 |

## 环境要求

- [Go](https://go.dev/dl/) >= 1.23
- [Node.js](https://nodejs.org/) >= 18
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 运行

```bash
# 安装依赖
cd frontend && npm install && cd ..

# 开发模式
wails dev -tags debug

# 构建生产版本
wails build
```

## 构建指定版本

```bash
wails build -ldflags "-X webhelper/global.Version=v1.0.0"
```

## 项目结构

```
├── api/                    # API 层 (HTTP/JS/Socket/Log)
├── global/                 # 全局配置、日志、语言包、版本号
├── lang/                   # 语言包 (go:embed 打包)
│   ├── default/            # 简体中文
│   ├── en/                 # English
│   ├── ja/                 # 日本語
│   └── ko/                 # 한국어
├── logger/                 # 日志系统
├── service/                # Wails 绑定方法
├── frontend/               # Vue 3 前端
│   └── src/components/     # 5 个功能页面
├── .github/workflows/      # CI/CD 自动构建
├── main.go                 # 程序入口
└── wails.json
```

## 添加语言

在 `lang/` 下新建目录，包含 `info.json` 和 `textmap.json`：

```json
// lang/ja/info.json
{
  "language_name": "日本語",
  "language_code": "ja-JP",
  "textmap_path": "textmap.json",
  "translation_progress": "100%",
  "translator": "",
  "last_updated": "",
  "version": "1.0.0"
}
```

重启后自动出现在语言切换列表中。

## CI/CD

推送 `v*` 标签自动构建多平台版本并发布 Release：

```bash
git tag v1.0.0
git push origin v1.0.0
```

构建目标：Windows (amd64)、Linux (amd64/arm64)、macOS (amd64/arm64)

## License

GPL-3.0
