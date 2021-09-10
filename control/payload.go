package control

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Payload map[string]json.RawMessage

func (p Payload) Action() string {
	action, _ := p.String("action")
	return action
}

func (p Payload) String(name string) (string, error) {
	result := ""
	err := json.Unmarshal(p[name], &result)
	return result, err
}

func (p Payload) Error() error {
	result, _ := p.String("result")
	if strings.ToLower(result) == "ok" || result == "" {
		return nil
	}
	return fmt.Errorf("result: %s :: %s", result, p.raw())
}

func (p Payload) raw() string {
	b, _ := json.Marshal(&p)
	return string(b)
}
