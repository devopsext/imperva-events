package output

import "git.exness.io/sre/imperva-events/pkg/common"

type Output interface {
	Send(event *common.Event) ([]byte, error)
}
