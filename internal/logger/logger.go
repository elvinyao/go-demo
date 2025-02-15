package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L 是全局可用的 SugaredLogger
var L *zap.SugaredLogger

// LogConfig 代表从 config.yaml 读取的日志相关配置
type LogConfig struct {
	Level       string `mapstructure:"level"`        // 日志级别: DEBUG, INFO, WARN, ERROR
	Output      string `mapstructure:"output"`       // 输出方式: "stdout", "file" 或 "both"
	Format      string `mapstructure:"format"`       // 日志格式: "json" 或 "console"
	Filename    string `mapstructure:"filename"`     // 日志文件名 (当 output=file/both 时)
	MaxBytes    int    `mapstructure:"max_bytes"`    // 文件最大大小(字节)，用于日志轮转
	BackupCount int    `mapstructure:"backup_count"` // 保留的旧文件个数
	MaxAgeDays  int    `mapstructure:"max_age_days"` // 文件最多保留天数
	Compress    bool   `mapstructure:"compress"`     // 是否压缩归档旧文件
}

// InitLogger 根据给定的 LogConfig 初始化 zap Logger
func InitLogger(cfg LogConfig) error {
	// 1. 解析日志级别
	var level zapcore.Level
	switch strings.ToUpper(cfg.Level) {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// 2. 构建 Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	// 可根据需要对时间格式或键名称进行修改
	// encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// encoderConfig.TimeKey = "timestamp"

	var encoder zapcore.Encoder
	switch strings.ToLower(cfg.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		// 默认为 console
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 3. 准备多个输出(cores)
	cores := make([]zapcore.Core, 0, 2)

	// 如果 output = file 或 both，需要创建文件写入器
	if cfg.Output == "file" || cfg.Output == "both" {
		// lumberjack 的 MaxSize 以 MB 为单位，需要把 max_bytes 转成 MB
		maxSizeMB := cfg.MaxBytes / (1024 * 1024)
		if maxSizeMB < 1 {
			maxSizeMB = 1 // 至少1MB
		}
		rotatingWriter := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    maxSizeMB,
			MaxBackups: cfg.BackupCount,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		}
		fileCore := zapcore.NewCore(encoder, zapcore.AddSync(rotatingWriter), level)
		cores = append(cores, fileCore)
	}

	// 如果 output = stdout 或 both，需要输出到控制台
	if cfg.Output == "stdout" || cfg.Output == "both" {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 如果既不是 file 也不是 stdout，就默认写到 stdout
	if cfg.Output != "file" && cfg.Output != "stdout" && cfg.Output != "both" {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	if len(cores) == 0 {
		return fmt.Errorf("no valid log output configured")
	}

	// 4. 合并 cores
	var combinedCore zapcore.Core
	if len(cores) == 1 {
		combinedCore = cores[0]
	} else {
		combinedCore = zapcore.NewTee(cores...)
	}

	// 5. 构建 Logger
	logger := zap.New(combinedCore, zap.AddCaller()) // 可选: zap.AddCaller() 显示调用位置
	L = logger.Sugar()

	return nil
}

// Sync 在程序退出时调用以刷新日志缓冲
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}
