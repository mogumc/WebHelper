package global

// Version 版本号，通过 ldflags 注入
// 编译时: wails build -ldflags "-X webhelper/global.Version=v1.2.3"
// 开发时默认为 "dev"
var Version = "dev"
