package imperva

import (
	"encoding/json"
	"fmt"
	"github.com/devopsext/imperva-events/pkg/common"
	"github.com/devopsext/imperva-events/pkg/output"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"sync"
	"time"
)

const apiInfraEvents = "https://my.imperva.com/api/v1/infra/events"
const apiAuditEvents = "https://api.imperva.com/audit-trail/v2/events"

// Traffic Statistics and Details API Get Infrastructure Protection Events
//curl --location --request POST 'https://my.imperva.com/api/v1/infra/events' \
//--header 'x-API-Key: YYY' \
//--header 'x-API-Id: XXX' \
//--header 'Content-Type: application/json'

// Audit Trail API
// curl -X 'GET' \
//  'https://api.imperva.com/audit-trail/v2/events?start=1660482470' \
//  -H 'accept: application/json' \
//  -H 'x-API-Id: YYY' \
//  -H 'x-API-Key: XXX'

type InfraEvent struct {
	EventTime     string `json:"eventTime"`
	EventType     string `json:"eventType"`
	BwTotal       int    `json:"bwTotal"`
	PpsTotal      int    `json:"ppsTotal"`
	BwPassed      int    `json:"bwPassed"`
	PpsPassed     int    `json:"ppsPassed"`
	BwBlocked     int    `json:"bwBlocked"`
	PpsBlocked    int    `json:"ppsBlocked"`
	EventTarget   string `json:"eventTarget"`
	ItemType      string `json:"itemType"`
	ReportedByPop string `json:"reportedByPop"`
}

type InfraEventsResponse struct {
	Events     []InfraEvent `json:"events"`
	Res        int          `json:"res"`
	ResMessage string       `json:"res_message"`
	DebugInfo  struct {
	} `json:"debug_info"`
}

type AuditEvent struct {
	Time            int64  `json:"time"`
	TypeKey         string `json:"type_key"`
	TypeDescription string `json:"type_description"`
	UserId          string `json:"user_id"`
	UserDetails     string `json:"user_details"`
	AccountId       string `json:"account_id"`
	ResourceTypeKey string `json:"resource_type_key"`
	ResourceId      string `json:"resource_id"`
	Message         string `json:"message"`
	ContextKey      string `json:"context_key"`
	AssumedByUser   string `json:"assumed_by_user"`
}

type AuditEventsErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Id      string `json:"id"`
}

type AuditEventsResponse struct {
	Total    int          `json:"total"`
	Elements []AuditEvent `json:"elements"`
}

type Imperva struct {
	client         *http.Client
	lastInfraEvent time.Time
	lastAuditEvent time.Time
	mutex          sync.RWMutex
	ticker         *time.Ticker
	id             string
	token          string
	accountId      string
	outputs        []output.Output
}

func (ie *InfraEvent) GetTime() time.Time {
	t, err := time.Parse("2006-01-02 15:04:05 UTC", ie.EventTime)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (ie *InfraEvent) String() string {
	return fmt.Sprintf("%s %s (POP: %s)", ie.EventType, ie.EventTarget, ie.ReportedByPop)
}

func (ae *AuditEvent) GetTime() time.Time {
	return time.Unix(ae.Time, 0)
}

func (ae *AuditEvent) String() string {
	//var m string
	//if len(ae.Message) > 40 {
	//	m = ae.Message[:40] + "..."
	//} else {
	//	m = ae.Message
	//}
	return fmt.Sprintf("%s by %s\n%s", ae.TypeDescription, ae.UserDetails, ae.Message)
}

func (i *Imperva) getInfraEvents() ([]InfraEvent, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	log.Debug().Msg("Getting new infra events")
	nle := i.lastInfraEvent
	var e []InfraEvent
	req, err := i.request(http.MethodPost, apiInfraEvents)
	if err != nil {
		return e, err
	}
	resp, err := i.client.Do(req)
	if err != nil {
		return e, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close body")
		}
	}(resp.Body)
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return e, err
	}
	var er InfraEventsResponse
	err = json.Unmarshal(b, &er)
	if err != nil {
		return e, err
	}
	for _, event := range er.Events {
		if event.GetTime().After(i.lastInfraEvent) {
			if event.GetTime().After(nle) {
				nle = event.GetTime()
			}
			e = append(e, event)
		}
	}
	i.lastInfraEvent = nle
	return e, nil
}

func (i *Imperva) getAuditEvents() ([]AuditEvent, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	log.Debug().Msg("Getting new audit events")
	nle := i.lastAuditEvent
	var e []AuditEvent
	req, err := i.request(http.MethodGet, apiAuditEvents)
	if err != nil {
		return e, err
	}
	q := req.URL.Query()
	q.Add("start", fmt.Sprintf("%d", i.lastAuditEvent.Unix()))
	req.URL.RawQuery = q.Encode()
	resp, err := i.client.Do(req)
	if err != nil {
		return e, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close body")
		}
	}(resp.Body)
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return e, err
	}
	if resp.StatusCode != http.StatusOK {
		var er AuditEventsErrorResponse
		err = json.Unmarshal(b, &er)
		if err != nil {
			return e, err
		}
		return e, fmt.Errorf("%s", er.Message)
	}
	var er AuditEventsResponse
	err = json.Unmarshal(b, &er)
	if err != nil {
		return e, err
	}
	for _, event := range er.Elements {
		if event.Time > i.lastAuditEvent.UnixMilli() {
			if event.Time > nle.UnixMilli() {
				nle = time.UnixMilli(event.Time)
			}
			e = append(e, event)
		}
	}
	i.lastAuditEvent = nle
	return e, nil
}

func (i *Imperva) AddOutput(o output.Output) {
	i.outputs = append(i.outputs, o)
}

func (ae *AuditEvent) toEvent() *common.Event {
	return &common.Event{
		Time:  time.UnixMilli(ae.Time),
		Title: "Audit",
		Body:  ae.String(),
	}
}

func (ie *InfraEvent) toEvent() *common.Event {
	return &common.Event{
		Time:  ie.GetTime(),
		Title: "Infra",
		Body:  ie.String(),
	}
}

func (i *Imperva) Send(e *common.Event) {
	for _, o := range i.outputs {
		b, err := o.Send(e)
		if err != nil {
			log.Error().Err(err).Msg("failed to send event")
		}
		if len(b) > 0 {
			log.Debug().Msg(string(b))
		}
	}
}

func (i *Imperva) Run(pollInterval int, wg *sync.WaitGroup) {
	i.ticker = time.NewTicker(time.Duration(pollInterval) * time.Second)
	go func() {
		for {
			<-i.ticker.C
			wg.Add(1)
			ies, err := i.getInfraEvents()
			if err != nil {
				log.Error().Err(err).Msg("failed to get new infra events")
			}
			for _, event := range ies {
				i.Send(event.toEvent())
			}

			aes, err := i.getAuditEvents()
			if err != nil {
				log.Error().Err(err).Msg("failed to get new audit events")
			}
			for _, event := range aes {
				i.Send(event.toEvent())
			}
			wg.Done()
		}
	}()
}

func (i *Imperva) request(method string, urlPath string) (*http.Request, error) {
	req, err := http.NewRequest(method, urlPath, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-API-Id", i.id)
	req.Header.Set("x-API-Key", i.token)
	req.Header.Set("accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("content-type", "application/json")
	}

	if i.accountId != "" {
		q := req.URL.Query()
		q.Add("account_id", i.accountId)
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func New(id string, token string, accountId string, initInterval int) (*Imperva, error) {
	i := &Imperva{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		id:             id,
		token:          token,
		accountId:      accountId,
		outputs:        append([]output.Output{}, output.NewStdout()),
		lastInfraEvent: time.Now().Add(time.Duration(-initInterval) * time.Minute),
		lastAuditEvent: time.Now().Add(time.Duration(-initInterval) * time.Minute),
	}

	return i, nil
}
