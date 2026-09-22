package rest

import (
	"bufio"
	"io"
	"iter"
	"strings"
)

// Event is one Server-Sent Event.
type Event struct {
	ID    string
	Event string
	Data  string
}

// Events parses a text/event-stream body. It closes body when iteration stops.
func Events(body io.ReadCloser) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {
		defer body.Close()
		sc := bufio.NewScanner(body)
		sc.Buffer(make([]byte, 64<<10), 4<<20)
		var ev Event
		var data []string
		for sc.Scan() {
			line := sc.Text()
			if line == "" {
				if len(data) > 0 || ev.Event != "" {
					ev.Data = strings.Join(data, "\n")
					if !yield(ev, nil) {
						return
					}
				}
				ev, data = Event{}, nil
				continue
			}
			if strings.HasPrefix(line, ":") {
				continue
			}
			field, value, _ := strings.Cut(line, ":")
			value = strings.TrimPrefix(value, " ")
			switch field {
			case "event":
				ev.Event = value
			case "data":
				data = append(data, value)
			case "id":
				ev.ID = value
			}
		}
		if err := sc.Err(); err != nil {
			yield(Event{}, err)
			return
		}
		if len(data) > 0 {
			ev.Data = strings.Join(data, "\n")
			yield(ev, nil)
		}
	}
}
