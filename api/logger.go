package api

import (
	"webhelper/global"
)

func LogInfo(msg string) {
	global.Log.Info(msg)
}

func LogWarn(msg string) {
	global.Log.Warn(msg)
}

func LogError(msg string) {
	global.Log.Error(msg)
}

func LogDebug(msg string) {
	global.Log.Debug(msg)
}