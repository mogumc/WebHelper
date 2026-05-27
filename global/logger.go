package global

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

// LogFormatter 控制台日志格式化器（带颜色）
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
	levelStr := strings.ToUpper(entry.Level.String())

	if IsDebug && entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		fmt.Fprintf(b, "\033[%dm[%s]\033[0m [%s] \033[%dm[%s]\033[0m %s %s: %s\n",
			colorGreen, processName, timestamp, levelColor, levelStr, funcVal, fileVal, entry.Message)
	} else {
		fmt.Fprintf(b, "\033[%dm[%s]\033[0m [%s] \033[%dm[%s]\033[0m %s\n",
			colorGreen, processName, timestamp, levelColor, levelStr, entry.Message)
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
	levelStr := strings.ToUpper(entry.Level.String())

	if IsDebug && entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		fmt.Fprintf(b, "[%s] [%s] [%s] %s %s: %s\n",
			processName, timestamp, levelStr, funcVal, fileVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] [%s] [%s] %s\n",
			processName, timestamp, levelStr, entry.Message)
	}
	return b.Bytes(), nil
}

var processName string

func init() {
	processName = filepath.Base(os.Args[0])
	processName = strings.TrimSuffix(processName, filepath.Ext(processName))
}

func GetProcessName() string {
	return processName
}

// SetLogLevel 动态设置日志等级
func SetLogLevel(level logrus.Level) {
	if Log != nil {
		Log.SetLevel(level)
	}
}

// InitLogger 初始化日志系统，每次运行时按启动时间保存日志文件，自动清理旧日志
func InitLogger() {
	Log = logrus.New()
	Log.SetFormatter(&LogFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	if IsDebug {
		Log.SetReportCaller(true)
	} else {
		Log.SetReportCaller(false)
	}

	logDir := GlobalConfig.LogDir
	if err := os.MkdirAll(logDir, 0755); err != nil {
		Log.Warnf("无法创建日志目录: %v", err)
	}

	cleanOldLogs(logDir, 10)

	startTime := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, startTime+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Warnf("无法打开日志文件: %v", err)
	} else {
		Log.SetOutput(os.Stdout)
		Log.AddHook(&FileHook{file: file})
	}
}

const colorGreen = 32

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

	var logFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry)
		}
	}

	if len(logFiles) <= maxKeep {
		return
	}

	// 文件名包含时间戳，按文件名排序即按时间排序
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Name() < logFiles[j].Name()
	})

	filesToDelete := logFiles[:len(logFiles)-maxKeep]
	for _, file := range filesToDelete {
		os.Remove(filepath.Join(logDir, file.Name()))
	}
}

func GetRuntimeInfo() string {
	return fmt.Sprintf("Go %s, OS %s, Arch %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
