package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

type Fields map[string]interface{}

type Logger struct {
	*zap.Logger
}

func New(env string) *Logger {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	core, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return &Logger{core}
}

func (l *Logger) WithFields(fields Fields) *Logger {
	zf := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zf = append(zf, zap.Any(k, v))
	}
	return &Logger{l.Logger.With(zf...)}
}

func (l *Logger) Info(msg string, fields ...Fields) {
	l.Logger.Info(msg, convert(fields)...)
}

func (l *Logger) Error(msg string, fields ...Fields) {
	l.Logger.Error(msg, convert(fields)...)
}

func (l *Logger) Fatal(msg string, fields ...Fields) {
	l.Logger.Fatal(msg, convert(fields)...)
}

func convert(fs []Fields) []zap.Field {
	if len(fs) == 0 {
		return nil
	}
	fields := make([]zap.Field, 0, len(fs))
	for _, f := range fs {
		for k, v := range f {
			fields = append(fields, zap.Any(k, v))
		}
	}
	return fields
}

func Field(k string, v interface{}) zap.Field {
	return zap.Any(k, v)
}

func Level(lv string) zapcore.Level {
	switch lv {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

type GormZapLogger struct {
	zapLogger *zap.SugaredLogger
	level     gormlogger.LogLevel
}

func NewGormZapWriter(logger *zap.SugaredLogger, level gormlogger.LogLevel) gormlogger.Writer {
	return &GormZapLogger{
		zapLogger: logger,
		level:     level,
	}
}

func (g *GormZapLogger) Printf(format string, args ...interface{}) {
	switch g.level {
	case gormlogger.Info:
		g.zapLogger.Infof(format, args...)
	case gormlogger.Warn:
		g.zapLogger.Warnf(format, args...)
	case gormlogger.Error:
		g.zapLogger.Errorf(format, args...)
	case gormlogger.Silent:
		// do nothing
	default:
		g.zapLogger.Infof(format, args...)
	}
}
