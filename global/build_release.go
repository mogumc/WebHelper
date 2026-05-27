//go:build !debug

package global

// IsDebug 标记是否为调试模式
// 编译后为 false，不允许使用 debug 模式
var IsDebug = false
