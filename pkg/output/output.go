package output

import "github.com/devopsext/imperva-events/pkg/common"

type Output interface {
	Send(event *common.Event) ([]byte, error)
}
