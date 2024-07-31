package logger

import (
	"cloud-collection/config"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var (
	defaultLogger  *logrus.Logger
	defaultLogPath = "/var/log/gse/cloud_collection.log"

	maxSize    = 10   // 每个日志文件最大10MB
	maxBackups = 3    // 保留最近的3个日志文件
	maxAge     = 7    // 保留最近7天的日志
	compress   = true // 是否压缩旧日志

	defaultLevel = logrus.InfoLevel
	LogLevelMap  = map[string]logrus.Level{
		"ERROR": logrus.ErrorLevel,
		"WARN":  logrus.WarnLevel,
		"INFO":  logrus.InfoLevel,
		"DEBUG": logrus.DebugLevel,
	}
)

func InitLogger(c config.Logger) {
	logger := logrus.New()
	// 设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 日志等级设置
	logger.SetLevel(defaultLevel)
	if c.Level != "" {
		if v, ok := LogLevelMap[c.Level]; ok {
			logger.SetLevel(v)
		}
	}

	// 设置日志路径
	if c.Path != "" {
		defaultLogPath = c.Path
	}

	// 确保日志目录存在
	logDir := filepath.Dir(defaultLogPath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			logger.Fatalf("Failed to create log directory: %v", err)
		}
	}
	// 检查文件路径是否可写
	if err := checkLogPathWritable(defaultLogPath); err != nil {
		logger.Fatalf("Log path is not writable: %v", err)
	}
	// 设置日志输出到文件，并使用 lumberjack 进行日志切分
	logger.SetOutput(&lumberjack.Logger{
		Filename:   defaultLogPath,
		MaxSize:    maxSize,    // 单个日志文件的最大尺寸（MB）
		MaxBackups: maxBackups, // 保留的旧日志文件的最大数量
		MaxAge:     maxAge,     // 保留的旧日志文件的最大天数
		Compress:   compress,   // 是否压缩旧日志文件
	})
	defaultLogger = logger
}

// 检查日志路径是否可写
func checkLogPathWritable(path string) error {
	// 当文件存在的时候 默认为可写
	if fileExists(path) {
		return nil
	}
	// 尝试通过创建文件进行判断是否可写
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()
	return os.Remove(path)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}
