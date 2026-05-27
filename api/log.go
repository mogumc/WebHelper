package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RequestLog 请求日志
type RequestLog struct {
	ID          int               `json:"id"`          // 序号
	Method      string            `json:"method"`      // 请求方法
	URL         string            `json:"url"`         // 请求地址
	Host        string            `json:"host"`        // 主机地址
	Path        string            `json:"path"`        // 请求路径
	Headers     map[string]string `json:"headers"`     // 请求头
	Cookies     map[string]string `json:"cookies"`     // Cookie
	BodyType    string            `json:"bodyType"`    // 数据类型
	Body        string            `json:"body"`        // 请求体
	StatusCode  int               `json:"statusCode"`  // 状态码
	Status      string            `json:"status"`      // 状态描述
	RespHeaders map[string]string `json:"respHeaders"` // 响应头
	RespBody    string            `json:"respBody"`    // 响应正文
	RespSize    int64             `json:"respSize"`    // 响应大小
	Time        string            `json:"time"`        // 请求耗时
	CreatedAt   string            `json:"createdAt"`   // 创建时间
}

// LogManager 日志管理器
type LogManager struct {
	logDir    string
	logs      []*RequestLog
	maxLogs   int
	fileName  string
}

var logManager *LogManager

// InitLogManager 初始化日志管理器
func InitLogManager(logDir string) {
	logManager = &LogManager{
		logDir:   logDir,
		logs:     make([]*RequestLog, 0),
		maxLogs:  1000,
		fileName: "http_requests.json",
	}
	logManager.loadLogs()
}

// GetLogManager 获取日志管理器
func GetLogManager() *LogManager {
	if logManager == nil {
		InitLogManager("logs")
	}
	return logManager
}

// loadLogs 从文件加载日志
func (m *LogManager) loadLogs() {
	filePath := filepath.Join(m.logDir, m.fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			m.logs = make([]*RequestLog, 0)
			return
		}
		return
	}

	var logs []*RequestLog
	if err := json.Unmarshal(data, &logs); err != nil {
		m.logs = make([]*RequestLog, 0)
		return
	}

	m.logs = logs
}

// saveLogs 保存日志到文件
func (m *LogManager) saveLogs() error {
	if err := os.MkdirAll(m.logDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(m.logDir, m.fileName)
	data, err := json.MarshalIndent(m.logs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// AddLog 添加日志
func (m *LogManager) AddLog(log *RequestLog) error {
	// 生成序号
	if len(m.logs) > 0 {
		log.ID = m.logs[0].ID + 1
	} else {
		log.ID = 1
	}

	log.CreatedAt = time.Now().Format(time.DateTime)

	// 插入到开头（最新的在前面）
	m.logs = append([]*RequestLog{log}, m.logs...)

	// 限制最大日志数量
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[:m.maxLogs]
	}

	return m.saveLogs()
}

// GetLogs 获取所有日志
func (m *LogManager) GetLogs() []*RequestLog {
	return m.logs
}

// GetLog 获取单条日志
func (m *LogManager) GetLog(id int) *RequestLog {
	for _, log := range m.logs {
		if log.ID == id {
			return log
		}
	}
	return nil
}

// DeleteLog 删除日志
func (m *LogManager) DeleteLog(id int) error {
	for i, log := range m.logs {
		if log.ID == id {
			m.logs = append(m.logs[:i], m.logs[i+1:]...)
			return m.saveLogs()
		}
	}
	return nil
}

// ClearLogs 清空日志
func (m *LogManager) ClearLogs() error {
	m.logs = make([]*RequestLog, 0)
	return m.saveLogs()
}

// SearchLogs 搜索日志
func (m *LogManager) SearchLogs(keyword string) []*RequestLog {
	var result []*RequestLog
	keyword = strings.ToLower(keyword)

	for _, log := range m.logs {
		if strings.Contains(strings.ToLower(log.URL), keyword) ||
			strings.Contains(strings.ToLower(log.Method), keyword) ||
			strings.Contains(strings.ToLower(log.Host), keyword) ||
			strings.Contains(strings.ToLower(log.Path), keyword) {
			result = append(result, log)
		}
	}

	return result
}

// CreateLogFromResponse 从响应创建日志
func CreateLogFromResponse(req *HttpRequest, resp *HttpResponse) *RequestLog {
	// 解析 URL
	host := ""
	path := req.URL
	if strings.HasPrefix(req.URL, "http://") {
		parts := strings.SplitN(strings.TrimPrefix(req.URL, "http://"), "/", 2)
		host = parts[0]
		if len(parts) > 1 {
			path = "/" + parts[1]
		}
	} else if strings.HasPrefix(req.URL, "https://") {
		parts := strings.SplitN(strings.TrimPrefix(req.URL, "https://"), "/", 2)
		host = parts[0]
		if len(parts) > 1 {
			path = "/" + parts[1]
		}
	}

	return &RequestLog{
		Method:      req.Method,
		URL:         req.URL,
		Host:        host,
		Path:        path,
		Headers:     req.Headers,
		Cookies:     req.Cookies,
		BodyType:    req.BodyType,
		Body:        req.Body,
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		RespHeaders: resp.Headers,
		RespBody:    resp.Body,
		RespSize:    resp.Size,
		Time:        resp.Time,
	}
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

// SortLogsByTime 按时间排序日志
func SortLogsByTime(logs []*RequestLog, desc bool) []*RequestLog {
	sorted := make([]*RequestLog, len(logs))
	copy(sorted, logs)

	sort.Slice(sorted, func(i, j int) bool {
		timeI, _ := time.Parse(time.DateTime, sorted[i].CreatedAt)
		timeJ, _ := time.Parse(time.DateTime, sorted[j].CreatedAt)
		if desc {
			return timeI.After(timeJ)
		}
		return timeI.Before(timeJ)
	})

	return sorted
}
