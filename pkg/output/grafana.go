package output

import (
	"github.com/devopsext/imperva-events/pkg/common"
	"github.com/devopsext/tools/vendors"
	"time"
)

type Grafana struct {
	client *vendors.Grafana
}

func (g *Grafana) Send(e *common.Event) ([]byte, error) {

	t := e.Time.Format(time.RFC3339Nano)

	return g.client.CreateCustomAnnotation(&vendors.GrafanaCreateAnnotationOptions{
		Time:    t,
		TimeEnd: t,
		Tags:    "imperva",
		Text:    e.Body,
	})
}

func NewGrafana(url string, apiKey string) *Grafana {
	return &Grafana{
		client: vendors.NewGrafana(vendors.GrafanaOptions{
			URL:    url,
			APIKey: apiKey,
		}),
	}
}
