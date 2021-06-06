package chatlog

import (
	"encoding/json"
	"fmt"
	"time"
)

/*
	https://logs.atitd.wiki/api/system/1623007144
	[{"message":"House Hyksos are in the final stages of electing a new Elder.","timestamp":1623007145}]
*/

type Timestamp struct {
	time.Time
}

func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
	var raw int64
	if err := json.Unmarshal(bytes, &raw); err != nil {
		fmt.Printf("error decoding timestamp: %s\n", err)
		return err
	}

	p.Time = time.Unix(raw, 0)
	return nil
}

type Message struct {
	Message   string
	Timestamp Timestamp
}
