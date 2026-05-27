package global

import "strings"

type GConfig struct {
	Language string `json:"language"`
	LogDir   string `json:"log_dir"`
	ProxyURL string `json:"proxy_url"` // 代理地址，支持 http:// 和 socks5://
	Timeout  int    `json:"timeout"`   // 默认请求超时（秒）
	LogLevel string `json:"log_level"` // 日志等级：debug/info/warn/error
}

var GlobalConfig = &GConfig{
	Language: "zh-CN",
	LogDir:   "logs",
	ProxyURL: "",
	Timeout:  30,
	LogLevel: "info",
}

// GetProxy 获取全局代理地址，未设置或为空时返回空字符串
func GetProxy() string {
	return NormalizeProxy(GlobalConfig.ProxyURL)
}

// NormalizeProxy 规范化代理地址：无协议前缀时默认 http://
func NormalizeProxy(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	// 已有协议前缀，直接返回
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") ||
		strings.HasPrefix(addr, "socks5://") || strings.HasPrefix(addr, "socks4://") {
		return addr
	}
	// 无前缀，默认 http://
	return "http://" + addr
}
