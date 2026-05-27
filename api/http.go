package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

// maxBodySize 限制响应体最大读取量为 50MB，防止 OOM
const maxBodySize = 50 << 20

// 默认 HTTP 客户端，复用连接池
var defaultHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	},
}

// HttpRequest 请求结构体
type HttpRequest struct {
	Method      string            `json:"method"`      // HTTP 方法
	URL         string            `json:"url"`         // 请求地址
	Headers     map[string]string `json:"headers"`     // 请求头
	Cookies     map[string]string `json:"cookies"`     // Cookie
	Proxy       string            `json:"proxy"`       // 代理地址
	BodyType    string            `json:"bodyType"`    // 数据类型: text, file, form, json
	Body        string            `json:"body"`        // 文本数据
	FilePath    string            `json:"filePath"`    // 文件路径
	ContentType string            `json:"contentType"` // Content-Type
	Timeout     int               `json:"timeout"`     // 超时时间(秒)
	Insecure    bool              `json:"insecure"`    // 忽略证书验证
}

// HttpResponse 响应结构体
type HttpResponse struct {
	StatusCode    int               `json:"statusCode"`    // 状态码
	Status        string            `json:"status"`        // 状态描述
	Headers       map[string]string `json:"headers"`       // 响应头
	Cookies       []CookieInfo      `json:"cookies"`       // 响应 Cookie
	Body          string            `json:"body"`          // 响应正文
	ContentType   string            `json:"contentType"`   // Content-Type
	ContentLength int64             `json:"contentLength"` // 内容长度
	Time          string            `json:"time"`          // 请求耗时
	Size          int64             `json:"size"`          // 响应大小
	Error         string            `json:"error"`         // 错误信息
}

// CookieInfo Cookie 信息
type CookieInfo struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Expires  string `json:"expires"`
	MaxAge   int    `json:"maxAge"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"httpOnly"`
	SameSite string `json:"sameSite"`
}

// FileData 文件数据
type FileData struct {
	FieldName   string `json:"fieldName"`   // 字段名
	FileName    string `json:"fileName"`    // 文件名
	ContentType string `json:"contentType"` // 文件类型
	Path        string `json:"path"`        // 文件路径
}

// normalizeURL 补全协议前缀
func normalizeURL(req *HttpRequest) {
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}
}

// buildHTTPClient 创建带代理和 TLS 配置的 HTTP 客户端
// 无自定义代理且不跳过验证时复用默认客户端
func buildHTTPClient(req *HttpRequest) *http.Client {
	// 无自定义设置时复用默认客户端（共享连接池）
	if req.Proxy == "" && !req.Insecure {
		if req.Timeout > 0 {
			return &http.Client{
				Timeout:   time.Duration(req.Timeout) * time.Second,
				Transport: defaultHTTPClient.Transport,
			}
		}
		return defaultHTTPClient
	}

	// 有自定义设置时创建新客户端
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	if req.Proxy != "" {
		proxyURL, err := url.Parse(req.Proxy)
		if err != nil {
			// 代理格式错误，降级为默认客户端
			return defaultHTTPClient
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: req.Insecure,
			},
		}
	} else if req.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}

// setupCookies 将请求中的 Cookie 设置到客户端 Jar
func setupCookies(client *http.Client, req *HttpRequest) {
	if len(req.Cookies) == 0 {
		return
	}
	cookieJar, _ := cookiejar.New(nil)
	var cookieList []*http.Cookie
	for name, value := range req.Cookies {
		cookieList = append(cookieList, &http.Cookie{Name: name, Value: value})
	}
	parsedURL, _ := url.Parse(req.URL)
	cookieJar.SetCookies(parsedURL, cookieList)
	client.Jar = cookieJar
}

// setRequestHeaders 设置自定义请求头（host 单独处理）
func setRequestHeaders(httpReq *http.Request, req *HttpRequest) {
	for key, value := range req.Headers {
		if strings.ToLower(key) == "host" {
			httpReq.Host = value
		} else {
			httpReq.Header.Set(key, value)
		}
	}
}

// parseHTTPResponse 从 http.Response 解析到 HttpResponse
func parseHTTPResponse(httpResp *http.Response, resp *HttpResponse, start time.Time) {
	resp.Time = time.Since(start).String()

	// 限制最大读取量防止 OOM
	body, err := io.ReadAll(io.LimitReader(httpResp.Body, maxBodySize))
	if err != nil {
		resp.Error = fmt.Sprintf("读取响应失败: %v", err)
		return
	}

	resp.StatusCode = httpResp.StatusCode
	resp.Status = httpResp.Status
	resp.Body = string(body)
	resp.ContentLength = httpResp.ContentLength
	resp.Size = int64(len(body))
	resp.ContentType = httpResp.Header.Get("Content-Type")

	resp.Headers = make(map[string]string)
	for key, values := range httpResp.Header {
		resp.Headers[key] = strings.Join(values, "; ")
	}

	resp.Cookies = convertCookies(httpResp.Cookies())
}

// SendRequest 发送 HTTP 请求
func SendRequest(req *HttpRequest) *HttpResponse {
	resp := &HttpResponse{}

	if req.URL == "" {
		resp.Error = "请求地址不能为空"
		return resp
	}

	normalizeURL(req)

	client := buildHTTPClient(req)

	// 构造请求体
	var bodyReader io.Reader
	var err error
	contentType := req.ContentType

	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		switch req.BodyType {
		case "file":
			if req.FilePath != "" {
				bodyReader, contentType, err = buildFileBody(req.FilePath)
				if err != nil {
					resp.Error = err.Error()
					return resp
				}
			}
		case "form":
			if req.Body != "" {
				bodyReader = strings.NewReader(req.Body)
				if contentType == "" {
					contentType = "application/x-www-form-urlencoded"
				}
			}
		case "json":
			if req.Body != "" {
				bodyReader = strings.NewReader(req.Body)
				if contentType == "" {
					contentType = "application/json"
				}
			}
		default:
			if req.Body != "" {
				bodyReader = strings.NewReader(req.Body)
				if contentType == "" {
					contentType = "text/plain"
				}
			}
		}
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		resp.Error = fmt.Sprintf("创建请求失败: %v", err)
		return resp
	}

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	setRequestHeaders(httpReq, req)
	setupCookies(client, req)

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		resp.Error = fmt.Sprintf("请求失败: %v", err)
		return resp
	}
	defer httpResp.Body.Close()

	parseHTTPResponse(httpResp, resp, start)
	return resp
}

// SendRequestWithFiles 带多文件上传的请求
func SendRequestWithFiles(req *HttpRequest, files []FileData) *HttpResponse {
	resp := &HttpResponse{}

	if req.URL == "" {
		resp.Error = "请求地址不能为空"
		return resp
	}

	normalizeURL(req)

	client := buildHTTPClient(req)

	// 构造 multipart 请求体
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if req.Body != "" {
		_ = writer.WriteField("text", req.Body)
	}

	for _, file := range files {
		f, err := os.Open(file.Path)
		if err != nil {
			resp.Error = fmt.Sprintf("打开文件 %s 失败: %v", file.Path, err)
			return resp
		}
		defer f.Close()

		part, err := writer.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			resp.Error = fmt.Sprintf("创建文件字段失败: %v", err)
			return resp
		}

		if _, err := io.Copy(part, f); err != nil {
			resp.Error = fmt.Sprintf("复制文件内容失败: %v", err)
			return resp
		}
	}

	writer.Close()

	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		resp.Error = fmt.Sprintf("创建请求失败: %v", err)
		return resp
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	setRequestHeaders(httpReq, req)
	setupCookies(client, req)

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		resp.Error = fmt.Sprintf("请求失败: %v", err)
		return resp
	}
	defer httpResp.Body.Close()

	parseHTTPResponse(httpResp, resp, start)
	return resp
}

// buildFileBody 构造单文件上传的 multipart 请求体
func buildFileBody(filePath string) (io.Reader, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("获取文件信息失败: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileInfo.Name())
	if err != nil {
		return nil, "", fmt.Errorf("创建表单文件失败: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, "", fmt.Errorf("复制文件内容失败: %v", err)
	}
	writer.Close()
	return body, writer.FormDataContentType(), nil
}

// ParseHeaders 解析请求头字符串
func ParseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	lines := strings.Split(headerStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}

// ParseCookies 解析 Cookie 字符串
func ParseCookies(cookieStr string) map[string]string {
	cookies := make(map[string]string)
	if cookieStr == "" {
		return cookies
	}

	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			cookies[name] = value
		}
	}

	return cookies
}

// convertCookies 将 http.Cookie 转换为 CookieInfo
func convertCookies(httpCookies []*http.Cookie) []CookieInfo {
	var cookies []CookieInfo
	for _, c := range httpCookies {
		sameSite := ""
		switch c.SameSite {
		case http.SameSiteStrictMode:
			sameSite = "Strict"
		case http.SameSiteLaxMode:
			sameSite = "Lax"
		case http.SameSiteNoneMode:
			sameSite = "None"
		}

		cookies = append(cookies, CookieInfo{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.Expires.Format(time.RFC1123),
			MaxAge:   c.MaxAge,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
			SameSite: sameSite,
		})
	}
	return cookies
}
