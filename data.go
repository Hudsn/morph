package morph

import (
	"encoding/json"
	"fmt"
)

type data struct {
	contents object
}

func newDataFromBytes(bytes []byte) (*data, error) {
	d := &data{}
	var raw interface{}
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	d.contents = d.toObject(raw)
	return d, nil
}

func (d *data) toObject(rawData interface{}) object {

	switch v := rawData.(type) {
	case float64: // all json nums are float64 for some reason? Go quirk maybe?
		return d.handleNumber(v)
	default:
		return &objectNull{}
	}
}

func (d *data) handleNumber(num float64) object {
	if num == float64(int64(num)) {
		return &objectInteger{value: int64(num)}
	} else {
		return &objectFloat{value: num}
	}
}
