//go:build debug

package global

// IsDebug 标记是否为调试模式
// 仅在使用 -tags debug 编译时为 true
var IsDebug = true
