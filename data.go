package morph

import "encoding/json"

type data struct {
	contents object
}

func newDataFromBytes(bytes []byte) (*data, error) {
	d := &data{}
	var raw interface{}
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *data) toObject(rawData interface{}) object {

	switch v := rawData.(type) {
	case float64: // all json nums are float64 for some reason? Go quirk maybe?
		return d.handleNumber(v)
	}

	return nil
}

func (d *data) handleNumber(num float64) object {
	if num == float64()
}
