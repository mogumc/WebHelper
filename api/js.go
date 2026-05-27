package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// JsExecResult JS 执行结果
type JsExecResult struct {
	Output   string `json:"output"`   // 控制台输出
	Result   string `json:"result"`   // 返回值
	Error    string `json:"error"`    // 错误信息
	Duration string `json:"duration"` // 执行耗时
	Success  bool   `json:"success"`  // 是否成功
}

// consoleLog 存储 console.log 输出
type consoleLogger struct {
	logs []string
}

func (c *consoleLogger) log(args ...interface{}) {
	c.logs = append(c.logs, fmt.Sprint(args...))
}

func (c *consoleLogger) logFn(call goja.FunctionCall) goja.Value {
	var args []interface{}
	for _, arg := range call.Arguments {
		args = append(args, arg.Export())
	}
	c.logs = append(c.logs, fmt.Sprint(args...))
	return goja.Undefined()
}

// ExecuteJs 执行 JavaScript 代码
func ExecuteJs(code string) *JsExecResult {
	result := &JsExecResult{
		Success: false,
	}

	if strings.TrimSpace(code) == "" {
		result.Error = "代码不能为空"
		return result
	}

	start := time.Now()

	// 创建 JS 虚拟机
	vm := goja.New()

	// 创建日志记录器
	logger := &consoleLogger{
		logs: make([]string, 0),
	}

	// 注入 console 对象
	consoleObj := vm.NewObject()
	consoleObj.Set("log", logger.logFn)
	consoleObj.Set("info", logger.logFn)
	consoleObj.Set("warn", logger.logFn)
	consoleObj.Set("error", logger.logFn)
	vm.Set("console", consoleObj)

	// 注入常用全局函数
	vm.Set("parseInt", parseInt)
	vm.Set("parseFloat", parseFloat)
	vm.Set("isNaN", isNaN)
	vm.Set("isFinite", isFinite)
	vm.Set("String", vm.Get("String"))
	vm.Set("Number", vm.Get("Number"))
	vm.Set("Boolean", vm.Get("Boolean"))
	vm.Set("Array", vm.Get("Array"))
	vm.Set("Object", vm.Get("Object"))
	vm.Set("JSON", vm.Get("JSON"))
	vm.Set("Math", vm.Get("Math"))
	vm.Set("Date", vm.Get("Date"))
	vm.Set("RegExp", vm.Get("RegExp"))

	// 执行代码
	val, err := vm.RunString(code)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start).String()
		result.Output = strings.Join(logger.logs, "\n")
		return result
	}

	// 获取返回值
	resultVal := ""
	if val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
		resultVal = fmt.Sprintf("%v", val.Export())
	}

	result.Output = strings.Join(logger.logs, "\n")
	result.Result = resultVal
	result.Duration = time.Since(start).String()
	result.Success = true

	return result
}

// parseInt 解析整数
func parseInt(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.NaN()
	}

	arg := call.Arguments[0].String()
	var radix int64 = 10
	if len(call.Arguments) > 1 {
		radix = call.Arguments[1].ToInteger()
	}

	var result int64
	var err error

	switch radix {
	case 16:
		_, err = fmt.Sscanf(arg, "%x", &result)
	case 8:
		_, err = fmt.Sscanf(arg, "%o", &result)
	default:
		_, err = fmt.Sscanf(arg, "%d", &result)
	}

	if err != nil {
		return goja.NaN()
	}

	return goja.New().ToValue(result)
}

// parseFloat 解析浮点数
func parseFloat(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.NaN()
	}

	arg := call.Arguments[0].String()
	var result float64
	_, err := fmt.Sscanf(arg, "%f", &result)
	if err != nil {
		return goja.NaN()
	}

	return goja.New().ToValue(result)
}

// isNaN 检查是否为 NaN
func isNaN(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.New().ToValue(true)
	}

	val := call.Arguments[0].ToNumber()
	if goja.IsNaN(val) {
		return goja.New().ToValue(true)
	}
	return goja.New().ToValue(false)
}

// isFinite 检查是否为有限数
func isFinite(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.New().ToValue(false)
	}

	val := call.Arguments[0].ToNumber()
	if goja.IsNaN(val) || goja.IsInfinity(val) {
		return goja.New().ToValue(false)
	}
	return goja.New().ToValue(true)
}
