package output

import (
	"git.exness.io/sre/imperva-events/pkg/common"
	"github.com/rs/zerolog/log"
)

type Stdout struct {
}

func (s *Stdout) Send(e *common.Event) ([]byte, error) {
	log.Info().Msg(e.String())
	return []byte{}, nil
}

func NewStdout() *Stdout {
	return &Stdout{}
}
