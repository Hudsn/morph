package morph

import (
	"encoding/json"
	"fmt"
)

// helpers for accessing data parts from arbitrary maps

func newObjectFromBytes(bytes []byte) (object, error) {
	// d
	var raw interface{}
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	obj := rawParseAny(raw)
	if objectIsError(obj) {
		return nil, fmt.Errorf("%s", obj.inspect())
	} // return d, nil
	return obj, nil
}

func rawParseAny(rawData interface{}) object {

	switch v := rawData.(type) {
	case float64: // all json nums are float64 for some reason? Go quirk maybe?
		return rawParseNumber(v)
	case bool:
		return rawParseBool(v)
	case map[string]interface{}:
		return rawParseMap(v)
	default:
		return &objectError{message: fmt.Sprintf("unable to read data into morph object: %+v", v)}
	}
}

func rawParseMap(m map[string]interface{}) object {
	ret := &objectMap{
		kvPairs: make(map[string]objectMapPair),
	}
	for k, v := range m {
		kvPair := objectMapPair{
			key:   k,
			value: rawParseAny(v),
		}
		ret.kvPairs[k] = kvPair
	}
	return ret
}

func rawParseBool(b bool) object {
	return &objectBoolean{value: b}
}

func rawParseNumber(num float64) object {
	if num == float64(int64(num)) {
		return &objectInteger{value: int64(num)}
	} else {
		return &objectFloat{value: num}
	}
}
