package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger(env string) {
	var encoderConfig zapcore.EncoderConfig
	var encoder zapcore.Encoder

	if env == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	Log = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Log)
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
