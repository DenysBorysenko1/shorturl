package audit

import "encoding/json"

type Event struct {
	TS      int64  `json:"ts"`
	Action  string `json:"action"`
	UserID  string `json:"user_id"`
	URL     string `json:"url"`
}

func (e *Event) Serialize() []byte {
	data, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return data
}