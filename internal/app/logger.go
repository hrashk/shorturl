package app

import (
	"os"

	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	logger zerolog.Logger
}

func NewZeroLogger() ZeroLogger {
	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return ZeroLogger{logger: zl}
}

func (zl ZeroLogger) Info(msg string, fields ...any) {
	zl.logger.Info().Fields(fields).Msg(msg)
}
