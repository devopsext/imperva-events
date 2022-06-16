package imperva

import (
	"encoding/json"
	"fmt"
	"git.exness.io/sre/imperva-events/pkg/common"
	"git.exness.io/sre/imperva-events/pkg/output"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://my.imperva.com"
const cmdEvents = "/api/v1/infra/events"

type ImpervaEvent struct {
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

type ImpervaEventsResponse struct {
	Events     []ImpervaEvent `json:"events"`
	Res        int            `json:"res"`
	ResMessage string         `json:"res_message"`
	DebugInfo  struct {
	} `json:"debug_info"`
}

type Imperva struct {
	client    *http.Client
	eventsReq *http.Request
	lastEvent time.Time
	mutex     sync.RWMutex
	ticker    *time.Ticker
	outputs   []output.Output
}

func (ie ImpervaEvent) GetTime() time.Time {
	t, err := time.Parse("2006-01-02 15:04:05 UTC", ie.EventTime)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (ie ImpervaEvent) String() string {
	return fmt.Sprintf("%s %s (POP: %s)", ie.EventType, ie.EventTarget, ie.ReportedByPop)
}

func (ie ImpervaEvent) toJson() ([]byte, error) {
	return json.Marshal(ie)
}

func (i *Imperva) GetNewEvents() ([]ImpervaEvent, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	log.Debug().Msg("Getting new events")
	nle := i.lastEvent
	var e []ImpervaEvent
	resp, err := i.client.Do(i.eventsReq)
	if err != nil {
		return e, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close body")
		}
	}(resp.Body)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return e, err
	}
	var er ImpervaEventsResponse
	err = json.Unmarshal(b, &er)
	if err != nil {
		return e, err
	}
	for _, event := range er.Events {
		if event.GetTime().After(i.lastEvent) {
			if event.GetTime().After(nle) {
				nle = event.GetTime()
			}
			e = append(e, event)
		}
	}
	i.lastEvent = nle
	return e, nil
}

func (i *Imperva) AddOutput(o output.Output) {
	i.outputs = append(i.outputs, o)
}

func (i *Imperva) Run(pollInterval int, initInterval int, wg *sync.WaitGroup) {
	i.lastEvent = time.Now().Add(time.Duration(-initInterval) * time.Minute)
	i.ticker = time.NewTicker(time.Duration(pollInterval) * time.Second)
	//defer i.ticker.Stop()
	go func() {
		for {
			<-i.ticker.C
			wg.Add(1)
			events, err := i.GetNewEvents()
			if err != nil {
				log.Error().Err(err).Msg("failed to get new events")
			}
			for _, event := range events {
				e := &common.Event{
					Time:  event.GetTime(),
					Title: "Imperva event",
					Body:  event.String(),
				}
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
			wg.Done()
		}
		wg.Done()
	}()
}

func NewImperva(id string, token string, accountId string) (*Imperva, error) {

	req, err := http.NewRequest("POST", baseURL+cmdEvents, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-API-Id", id)
	req.Header.Set("x-API-Key", token)
	q := req.URL.Query()
	q.Add("account_id", accountId)
	req.URL.RawQuery = q.Encode()

	i := &Imperva{
		client:    &http.Client{},
		eventsReq: req,
		outputs:   append([]output.Output{}, output.NewStdout()),
	}

	return i, nil
}
