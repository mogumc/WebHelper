package logger

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// ANSI 颜色代码
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

// LogFormatter 自定义日志格式化器
type LogFormatter struct{}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		fmt.Fprintf(b, "%s[%s] \033[%dm[%s]\033[0m %s %s: %s\n",
			"[Logger]", timestamp, levelColor, entry.Level, funcVal, fileVal, entry.Message)
	} else {
		fmt.Fprintf(b, "%s[%s] \033[%dm[%s]\033[0m %s\n",
			"[Logger]", timestamp, levelColor, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}

// FileFormatter 文件日志格式化器（不包含 ANSI 颜色代码）
type FileFormatter struct{}

func (f *FileFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		fmt.Fprintf(b, "%s[%s] [%s] %s %s: %s\n",
			"[Logger]", timestamp, entry.Level, funcVal, fileVal, entry.Message)
	} else {
		fmt.Fprintf(b, "%s[%s] [%s] %s\n",
			"[Logger]", timestamp, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}

// InitLogger 初始化日志系统
// 每次运行时按照启动时间保存日志文件，并自动删除最新十个日志之前的日志
func InitLogger() {
	Log = logrus.New()
	Log.SetReportCaller(true)
	Log.SetFormatter(&LogFormatter{})
	Log.SetLevel(logrus.DebugLevel)

	// 确保日志目录存在
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		Log.Warnf("无法创建日志目录: %v", err)
	}

	// 清理旧日志（保留最新10个）
	cleanOldLogs(logDir, 10)

	// 创建按启动时间命名的日志文件
	startTime := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, startTime+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Warnf("无法打开日志文件: %v", err)
	} else {
		// 控制台输出使用带颜色的格式
		Log.SetOutput(os.Stdout)

		// 文件输出使用无颜色的格式，通过 Hook 实现
		Log.AddHook(&FileHook{file: file})
	}
}

// FileHook 将日志同时写入文件的 Hook
type FileHook struct {
	file *os.File
}

func (h *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *FileHook) Fire(entry *logrus.Entry) error {
	formatter := &FileFormatter{}
	line, err := formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.file.Write(line)
	return err
}

// cleanOldLogs 清理旧日志文件，保留最新的 maxKeep 个日志
func cleanOldLogs(logDir string, maxKeep int) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	// 收集所有 .log 文件
	var logFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry)
		}
	}

	// 如果日志文件数量未超过限制，无需清理
	if len(logFiles) <= maxKeep {
		return
	}

	// 按文件名排序（文件名包含时间戳，排序即按时间排序）
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Name() < logFiles[j].Name()
	})

	// 删除超出限制的旧日志
	filesToDelete := logFiles[:len(logFiles)-maxKeep]
	for _, file := range filesToDelete {
		filePath := filepath.Join(logDir, file.Name())
		os.Remove(filePath)
	}
}
