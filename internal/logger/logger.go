package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log        *zap.SugaredLogger
	atomicLevel zap.AtomicLevel
)

// SetLevel 动态更新日志级别
func SetLevel(level string) error {
	switch level {
	case "debug":
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case "warn":
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	default:
		atomicLevel.SetLevel(zapcore.InfoLevel)
	}
	return nil
}

func Init(level string) error {
	atomicLevel = parseLevel(level)
	var config zap.Config

	config = zap.Config{
		Level:            atomicLevel,
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger.Sugar()
	return nil
}

func parseLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Debugf(template string, args ...interface{}) {
	Log.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Log.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Log.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Log.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	Log.Fatalf(template, args...)
}
