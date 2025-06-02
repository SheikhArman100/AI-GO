package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	AppLogger   *zap.Logger // For application events (server/db start/stop)
	ErrorLogger *zap.Logger // For API errors
	QueryLogger *zap.Logger // For SQL queries
)

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "msg"
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = customLevelEncoder
	encoderConfig.CallerKey = ""
	encoderConfig.NameKey = ""
	encoderConfig.StacktraceKey = ""

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(level.String())
}

func InitLoggers() error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	currentDate := time.Now().Format("2006-01-02")

	// Create the core for app logger
	appLogFile, err := os.OpenFile(
		filepath.Join(logDir, fmt.Sprintf("app_%s.log", currentDate)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	core := zapcore.NewCore(
		getEncoder(),
		zapcore.AddSync(appLogFile),
		zap.InfoLevel,
	)

	AppLogger = zap.New(core)

	// Initialize error logger
	errorLogFile, err := os.OpenFile(
		filepath.Join(logDir, fmt.Sprintf("error_%s.log", currentDate)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	errorCore := zapcore.NewCore(
		getEncoder(),
		zapcore.AddSync(errorLogFile),
		zap.ErrorLevel,
	)

	ErrorLogger = zap.New(errorCore)

	// Initialize query logger
	queryLogFile, err := os.OpenFile(
		filepath.Join(logDir, fmt.Sprintf("query_%s.log", currentDate)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	queryCore := zapcore.NewCore(
		getEncoder(),
		zapcore.AddSync(queryLogFile),
		zap.InfoLevel,
	)

	QueryLogger = zap.New(queryCore)

	return nil
}
