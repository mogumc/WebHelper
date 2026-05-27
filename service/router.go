package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"webhelper/api"
	"webhelper/global"

	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetLangTextMap() map[string]string {
	return global.GetLangTextMap()
}

func (a *App) GetLangPack() *global.LanguagePack {
	langPack, err := global.GetLangPack()
	if err != nil {
		global.Log.Warnf("获取语言包失败: %v", err)
		return nil
	}
	return langPack
}

func (a *App) GetALLLang() []global.LanguageInfo {
	return global.GetLangInfoList()
}

func (a *App) SetLanguage(langCode string) bool {
	global.GlobalConfig.Language = langCode
	global.ClearLangCache()
	global.UpdateCurrentLangPath()
	return true
}

func (a *App) GetCurrentLang() string {
	return global.GlobalConfig.Language
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	URL string `json:"url"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Timeout int `json:"timeout"`
}

// GetProxy 获取全局代理配置
func (a *App) GetProxy() ProxyConfig {
	return ProxyConfig{
		URL: global.GlobalConfig.ProxyURL,
	}
}

// SetProxy 设置全局代理配置
func (a *App) SetProxy(url string) bool {
	global.GlobalConfig.ProxyURL = url
	global.Log.Infof("全局代理已更新: %s", url)
	return true
}

// GetProxyAddress 获取全局代理地址字符串（已规范化），供 HTTP 请求使用
func (a *App) GetProxyAddress() string {
	return global.GetProxy()
}

// GetTimeout 获取默认超时设置
func (a *App) GetTimeout() TimeoutConfig {
	return TimeoutConfig{
		Timeout: global.GlobalConfig.Timeout,
	}
}

// SetTimeout 设置默认超时
func (a *App) SetTimeout(timeout int) bool {
	if timeout < 1 {
		timeout = 30
	}
	global.GlobalConfig.Timeout = timeout
	global.Log.Infof("默认超时已更新: %d", timeout)
	return true
}

func (a *App) GetLogFiles() []string {
	logDir := global.GlobalConfig.LogDir
	entries, err := os.ReadDir(logDir)
	if err != nil {
		global.Log.Warnf("读取日志目录失败: %v", err)
		return []string{}
	}

	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(logFiles)))
	return logFiles
}

func (a *App) GetLogFileContent(filename string) string {
	// 安全检查：防止路径遍历
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		global.Log.Warnf("非法的日志文件名: %s", filename)
		return ""
	}

	logPath := filepath.Join(global.GlobalConfig.LogDir, filename)
	data, err := os.ReadFile(logPath)
	if err != nil {
		global.Log.Warnf("读取日志文件失败: %v", err)
		return ""
	}
	return string(data)
}

func (a *App) SetLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug":
		global.SetLogLevel(logrus.DebugLevel)
		global.GlobalConfig.LogLevel = "debug"
	case "info":
		global.SetLogLevel(logrus.InfoLevel)
		global.GlobalConfig.LogLevel = "info"
	case "warn":
		global.SetLogLevel(logrus.WarnLevel)
		global.GlobalConfig.LogLevel = "warn"
	case "error":
		global.SetLogLevel(logrus.ErrorLevel)
		global.GlobalConfig.LogLevel = "error"
	default:
		global.Log.Warnf("未知的日志等级: %s", level)
		return false
	}
	global.Log.Infof("日志等级已切换为: %s", strings.ToUpper(global.GlobalConfig.LogLevel))
	return true
}

func (a *App) GetLogLevel() string {
	return global.GlobalConfig.LogLevel
}

func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) WindowToggleMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) WindowClose() {
	runtime.Quit(a.ctx)
}

func (a *App) GetProcessName() string {
	return global.GetProcessName()
}

func (a *App) GetVersion() string {
	return global.Version
}

func (a *App) OpenFileSelect() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "所有文件", Pattern: "*.*"},
			{DisplayName: "文本文件", Pattern: "*.txt"},
			{DisplayName: "JSON 文件", Pattern: "*.json"},
		},
	})
	if err != nil {
		global.Log.Warnf("打开文件对话框失败: %v", err)
		return ""
	}
	return file
}

func (a *App) OpenFolderSelect() string {
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择目录",
	})
	if err != nil {
		global.Log.Warnf("打开目录对话框失败: %v", err)
		return ""
	}
	return folder
}

func (a *App) SaveFileSelect() string {
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "保存文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "文本文件", Pattern: "*.txt"},
			{DisplayName: "JSON 文件", Pattern: "*.json"},
		},
	})
	if err != nil {
		global.Log.Warnf("打开保存对话框失败: %v", err)
		return ""
	}
	return file
}

func (a *App) ReadFileContent(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		global.Log.Warnf("读取文件失败: %v", err)
		return fmt.Sprintf("读取失败: %v", err)
	}
	return string(data)
}

func (a *App) WriteFileContent(path string, content string) bool {
	if path == "" {
		return false
	}
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		global.Log.Warnf("写入文件失败: %v", err)
		return false
	}
	global.Log.Infof("文件写入成功: %s", path)
	return true
}

func (a *App) Notify(title string, message string) {
	runtime.EventsEmit(a.ctx, "notification", map[string]string{
		"title":   title,
		"message": message,
	})
}

// SendHttpRequest 发送 HTTP 请求
func (a *App) SendHttpRequest(method, url, headers, cookies, proxy, bodyType, body, filePath, contentType string, timeout int, insecure bool, saveLog bool) *api.HttpResponse {
	// 解析请求头
	headerMap := api.ParseHeaders(headers)

	// 解析 Cookie
	cookieMap := api.ParseCookies(cookies)

	// 代理优先级：请求级代理 > 全局代理
	requestProxy := global.NormalizeProxy(proxy)
	if requestProxy == "" {
		requestProxy = global.GetProxy()
	}

	// 超时优先级：请求级 > 全局默认
	if timeout <= 0 {
		timeout = global.GlobalConfig.Timeout
	}

	// 日志保存优先级：显式传入 > 全局默认
	// saveLog 由前端传入，这里不覆盖

	req := &api.HttpRequest{
		Method:      method,
		URL:         url,
		Headers:     headerMap,
		Cookies:     cookieMap,
		Proxy:       requestProxy,
		BodyType:    bodyType,
		Body:        body,
		FilePath:    filePath,
		ContentType: contentType,
		Timeout:     timeout,
		Insecure:    insecure,
	}

	global.Log.Infof("发送 HTTP 请求: %s %s", method, url)
	resp := api.SendRequest(req)

	if resp.Error != "" {
		global.Log.Warnf("HTTP 请求失败: %s", resp.Error)
	} else {
		global.Log.Infof("HTTP 请求完成: %d %s (%v)", resp.StatusCode, resp.Status, resp.Time)
	}

	// 保存日志
	if saveLog {
		logManager := api.GetLogManager()
		requestLog := api.CreateLogFromResponse(req, resp)
		if err := logManager.AddLog(requestLog); err != nil {
			global.Log.Warnf("保存请求日志失败: %v", err)
		}
	}

	return resp
}

// GetRequestLogs 获取请求日志列表
func (a *App) GetRequestLogs() []*api.RequestLog {
	logManager := api.GetLogManager()
	return logManager.GetLogs()
}

// GetRequestLog 获取单条请求日志
func (a *App) GetRequestLog(id int) *api.RequestLog {
	logManager := api.GetLogManager()
	return logManager.GetLog(id)
}

// DeleteRequestLog 删除请求日志
func (a *App) DeleteRequestLog(id int) bool {
	logManager := api.GetLogManager()
	if err := logManager.DeleteLog(id); err != nil {
		global.Log.Warnf("删除请求日志失败: %v", err)
		return false
	}
	return true
}

// ClearRequestLogs 清空请求日志
func (a *App) ClearRequestLogs() bool {
	logManager := api.GetLogManager()
	if err := logManager.ClearLogs(); err != nil {
		global.Log.Warnf("清空请求日志失败: %v", err)
		return false
	}
	return true
}

// SearchRequestLogs 搜索请求日志
func (a *App) SearchRequestLogs(keyword string) []*api.RequestLog {
	logManager := api.GetLogManager()
	return logManager.SearchLogs(keyword)
}

// ExecuteJs 执行 JavaScript 代码
func (a *App) ExecuteJs(code string) *api.JsExecResult {
	global.Log.Infof("执行 JavaScript 代码，长度: %d", len(code))
	result := api.ExecuteJs(code)

	if result.Success {
		global.Log.Infof("JavaScript 执行成功，耗时: %s", result.Duration)
	} else {
		global.Log.Warnf("JavaScript 执行失败: %s", result.Error)
	}

	return result
}

// ConnectSocket 连接 WebSocket 或 TCP
func (a *App) ConnectSocket(protocol, host string, port int, path string, headers string) bool {
	// 确保 path 以 / 开头
	if path == "" {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}

	var url string
	if protocol == "tcp" {
		url = fmt.Sprintf("%s:%d", host, port)
	} else {
		scheme := "ws"
		if protocol == "wss" {
			scheme = "wss"
		}
		url = fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
	}

	global.Log.Infof("连接 %s: %s", protocol, url)

	// 发送连接中事件
	runtime.EventsEmit(a.ctx, "socket-status", map[string]string{
		"type":    "connecting",
		"content": "正在连接...",
	})

	if protocol == "tcp" {
		// TCP 连接
		config := &api.TCPConfig{
			Host:  host,
			Port:  port,
			Proxy: global.GetProxy(),
		}

		err := api.ConnectTCP(config,
			// onMessage
			func(message string) {
				runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
					"type":    "received",
					"content": message,
				})
			},
			// onError
			func(err error) {
				// 连接失败时的错误回调
				runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
					"type":    "error",
					"content": err.Error(),
				})
			},
			// onClose
			func() {
				runtime.EventsEmit(a.ctx, "socket-status", map[string]string{
					"type":    "disconnected",
					"content": "连接已断开",
				})
			},
		)

		if err != nil {
			global.Log.Warnf("TCP 连接失败: %v", err)
			runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
				"type":    "error",
				"content": "连接失败: " + err.Error(),
			})
			return false
		}
	} else {
		// WebSocket 连接
		headerMap := api.ParseHeaders(headers)
		config := &api.WebSocketConfig{
			URL:     url,
			Headers: headerMap,
			Proxy:   global.GetProxy(),
		}

		err := api.ConnectWebSocket(config,
			// onMessage
			func(message string) {
				runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
					"type":    "received",
					"content": message,
				})
			},
			// onError
			func(err error) {
				// 连接失败时的错误回调
				runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
					"type":    "error",
					"content": err.Error(),
				})
			},
			// onClose
			func() {
				runtime.EventsEmit(a.ctx, "socket-status", map[string]string{
					"type":    "disconnected",
					"content": "连接已断开",
				})
			},
		)

		if err != nil {
			global.Log.Warnf("WebSocket 连接失败: %v", err)
			runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
				"type":    "error",
				"content": "连接失败: " + err.Error(),
			})
			return false
		}
	}

	global.Log.Infof("连接成功: %s", url)
	runtime.EventsEmit(a.ctx, "socket-status", map[string]string{
		"type":    "connected",
		"content": "连接成功",
	})

	return true
}

// SendSocketMessage 发送 Socket 消息
func (a *App) SendSocketMessage(protocol, message string) bool {
	global.Log.Infof("发送 %s 消息: %s", protocol, message)

	var err error
	if protocol == "tcp" {
		err = api.SendTCPMessage(message)
	} else {
		err = api.SendWebSocketMessage(message)
	}

	if err != nil {
		global.Log.Warnf("发送消息失败: %v", err)
		runtime.EventsEmit(a.ctx, "socket-message", map[string]string{
			"type":    "error",
			"content": "发送失败: " + err.Error(),
		})
		return false
	}

	// 发送成功事件
	runtime.EventsEmit(a.ctx, "socket-sent", map[string]string{
		"type":    "sent",
		"content": message,
	})

	return true
}

// DisconnectSocket 断开 Socket 连接
func (a *App) DisconnectSocket(protocol string) bool {
	global.Log.Infof("断开 %s 连接", protocol)

	var err error
	if protocol == "tcp" {
		err = api.CloseTCP()
	} else {
		err = api.CloseWebSocket()
	}

	if err != nil {
		global.Log.Warnf("断开连接失败: %v", err)
		return false
	}

	runtime.EventsEmit(a.ctx, "socket-status", map[string]string{
		"type":    "disconnected",
		"content": "已断开连接",
	})

	return true
}

// IsSocketConnected 检查 Socket 是否已连接
func (a *App) IsSocketConnected(protocol string) bool {
	if protocol == "tcp" {
		return api.IsTCPConnected()
	}
	return api.IsWebSocketConnected()
}
