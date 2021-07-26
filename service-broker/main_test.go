package broker

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logBuffer = &bytes.Buffer{}

func TestMain(m *testing.M) {
	// logger io.Writer to buffer to disable display log after terminate server
	log.Logger = zerolog.New(logBuffer).With().Timestamp().Logger()

	m.Run()
}
