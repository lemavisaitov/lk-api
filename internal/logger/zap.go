package logger

import (
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	log *zap.Logger
}

var global zapLogger

func Debug(msg string, fields ...zapcore.Field) {
	global.log.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	global.log.Info(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	global.log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	global.log.Fatal(msg, fields...)
}

func Stop() error {
	return global.log.Sync()
}

func Init(logLevel string) error {
	// Настраиваем ротацию логов
	fileSync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/song-library.log", // Все логи пишем в один файл
		MaxSize:    10,                      // Максимальный размер файла (в МБ)
		MaxBackups: 5,                       // Сколько старых файлов хранить
		MaxAge:     30,                      // Сколько дней хранить файлы
		Compress:   true,                    // Сжимать старые файлы (gzip)
	})

	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	cEncoder := zapcore.NewConsoleEncoder(cfg) // вывод в консоль
	fEncoder := zapcore.NewConsoleEncoder(cfg) // вывод в файл

	lvl, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return errors.Wrap(err, "failed to parse log level")
	}
	core := zapcore.NewTee(
		zapcore.NewCore(cEncoder, zapcore.AddSync(os.Stdout), lvl),
		zapcore.NewCore(fEncoder, fileSync, lvl),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	global = zapLogger{log: log}
	return nil
}
