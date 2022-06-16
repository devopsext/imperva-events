package common

import "time"

type Event struct {
	Time  time.Time `json:"time"`
	Title string    `json:"title"`
	Body  string    `json:"body"`
}

func (e *Event) String() string {
	return e.Title + ": " + e.Body
}
