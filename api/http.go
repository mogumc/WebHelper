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

// SendRequest 发送 HTTP 请求
func SendRequest(req *HttpRequest) *HttpResponse {
	resp := &HttpResponse{
		Headers: make(map[string]string),
	}

	if req.URL == "" {
		resp.Error = "请求地址不能为空"
		return resp
	}

	// 补全协议
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}

	// 设置超时
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: timeout,
	}

	// 配置代理
	if req.Proxy != "" {
		proxyURL, err := url.Parse(req.Proxy)
		if err != nil {
			resp.Error = fmt.Sprintf("代理地址格式错误: %v", err)
			return resp
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: req.Insecure,
			},
		}
	} else if req.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	// 创建请求体
	var bodyReader io.Reader
	contentType := req.ContentType

	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		switch req.BodyType {
		case "file":
			// 文件上传
			if req.FilePath != "" {
				file, err := os.Open(req.FilePath)
				if err != nil {
					resp.Error = fmt.Sprintf("打开文件失败: %v", err)
					return resp
				}
				defer file.Close()

				// 获取文件信息
				fileInfo, err := file.Stat()
				if err != nil {
					resp.Error = fmt.Sprintf("获取文件信息失败: %v", err)
					return resp
				}

				// 使用 multipart 上传
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, err := writer.CreateFormFile("file", fileInfo.Name())
				if err != nil {
					resp.Error = fmt.Sprintf("创建表单文件失败: %v", err)
					return resp
				}
				if _, err := io.Copy(part, file); err != nil {
					resp.Error = fmt.Sprintf("复制文件内容失败: %v", err)
					return resp
				}
				writer.Close()
				bodyReader = body
				contentType = writer.FormDataContentType()
			}

		case "form":
			// 表单数据
			if req.Body != "" {
				bodyReader = strings.NewReader(req.Body)
				if contentType == "" {
					contentType = "application/x-www-form-urlencoded"
				}
			}

		case "json":
			// JSON 数据
			if req.Body != "" {
				bodyReader = strings.NewReader(req.Body)
				if contentType == "" {
					contentType = "application/json"
				}
			}

		default:
			// 文本数据
			if req.Body != "" {
				bodyReader = strings.NewReader(req.Body)
				if contentType == "" {
					contentType = "text/plain"
				}
			}
		}
	}

	// 创建请求
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		resp.Error = fmt.Sprintf("创建请求失败: %v", err)
		return resp
	}

	// 设置 Content-Type
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// 设置自定义请求头（需要单独处理 host）
	for key, value := range req.Headers {
		if strings.ToLower(key) == "host" {
			httpReq.Host = value
		} else {
			httpReq.Header.Set(key, value)
		}
	}

	// 设置 Cookie
	if len(req.Cookies) > 0 {
		cookieJar, _ := cookiejar.New(nil)
		var cookieList []*http.Cookie
		for name, value := range req.Cookies {
			cookieList = append(cookieList, &http.Cookie{
				Name:  name,
				Value: value,
			})
		}
		parsedURL, _ := url.Parse(req.URL)
		cookieJar.SetCookies(parsedURL, cookieList)
		client.Jar = cookieJar
	}

	// 记录开始时间
	start := time.Now()

	// 发送请求
	httpResp, err := client.Do(httpReq)
	if err != nil {
		resp.Error = fmt.Sprintf("请求失败: %v", err)
		return resp
	}
	defer httpResp.Body.Close()

	// 记录耗时
	resp.Time = time.Since(start).String()

	// 读取响应体
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Error = fmt.Sprintf("读取响应失败: %v", err)
		return resp
	}

	// 设置响应
	resp.StatusCode = httpResp.StatusCode
	resp.Status = httpResp.Status
	resp.Body = string(body)
	resp.ContentLength = httpResp.ContentLength
	resp.Size = int64(len(body))

	// 获取 Content-Type
	resp.ContentType = httpResp.Header.Get("Content-Type")

	// 获取响应头
	for key, values := range httpResp.Header {
		resp.Headers[key] = strings.Join(values, "; ")
	}

	// 获取响应 Cookie
	resp.Cookies = convertCookies(httpResp.Cookies())

	return resp
}

// SendRequestWithFiles 带文件的请求
func SendRequestWithFiles(req *HttpRequest, files []FileData) *HttpResponse {
	resp := &HttpResponse{
		Headers: make(map[string]string),
	}

	if req.URL == "" {
		resp.Error = "请求地址不能为空"
		return resp
	}

	// 补全协议
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}

	// 设置超时
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: timeout,
	}

	// 配置代理
	if req.Proxy != "" {
		proxyURL, err := url.Parse(req.Proxy)
		if err != nil {
			resp.Error = fmt.Sprintf("代理地址格式错误: %v", err)
			return resp
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: req.Insecure,
			},
		}
	} else if req.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	// 创建 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文本字段
	if req.Body != "" {
		_ = writer.WriteField("text", req.Body)
	}

	// 添加文件
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

	// 创建请求
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		resp.Error = fmt.Sprintf("创建请求失败: %v", err)
		return resp
	}

	// 设置 Content-Type
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 设置自定义请求头（需要单独处理 host）
	for key, value := range req.Headers {
		if strings.ToLower(key) == "host" {
			httpReq.Host = value
		} else {
			httpReq.Header.Set(key, value)
		}
	}

	// 设置 Cookie
	if len(req.Cookies) > 0 {
		cookieJar, _ := cookiejar.New(nil)
		var cookieList []*http.Cookie
		for name, value := range req.Cookies {
			cookieList = append(cookieList, &http.Cookie{
				Name:  name,
				Value: value,
			})
		}
		parsedURL, _ := url.Parse(req.URL)
		cookieJar.SetCookies(parsedURL, cookieList)
		client.Jar = cookieJar
	}

	// 记录开始时间
	start := time.Now()

	// 发送请求
	httpResp, err := client.Do(httpReq)
	if err != nil {
		resp.Error = fmt.Sprintf("请求失败: %v", err)
		return resp
	}
	defer httpResp.Body.Close()

	// 记录耗时
	resp.Time = time.Since(start).String()

	// 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Error = fmt.Sprintf("读取响应失败: %v", err)
		return resp
	}

	// 设置响应
	resp.StatusCode = httpResp.StatusCode
	resp.Status = httpResp.Status
	resp.Body = string(respBody)
	resp.ContentLength = httpResp.ContentLength
	resp.Size = int64(len(respBody))

	// 获取 Content-Type
	resp.ContentType = httpResp.Header.Get("Content-Type")

	// 获取响应头
	for key, values := range httpResp.Header {
		resp.Headers[key] = strings.Join(values, "; ")
	}

	// 获取响应 Cookie
	resp.Cookies = convertCookies(httpResp.Cookies())

	return resp
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
