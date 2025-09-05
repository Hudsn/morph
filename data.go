package morph

import (
	"encoding/json"
	"fmt"
)

// helpers for accessing data parts from arbitrary maps

func newObjectFromAny(t interface{}) (object, error) {
	return rawParseAny(t, false)
}

func newObjectFromBytes(bytes []byte) (object, error) {
	// d
	var raw interface{}
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	obj, err := rawParseAnyJSON(raw)
	if err != nil {
		return obj_global_null, err
	}
	return obj, nil
}

func rawParseAnyJSON(rawData interface{}) (object, error) {
	switch v := rawData.(type) {
	case float64:
		return rawParseNumJson(v), nil
	default:
		return rawParseAny(v, true)
	}
}

func rawParseAny(rawData interface{}, isJSON bool) (object, error) {

	switch v := rawData.(type) {
	case int, int16, int32, int64, float32, float64:
		return rawParseNum(v, isJSON)
	case bool:
		return &objectBoolean{value: v}, nil
	case string:
		return &objectString{value: v}, nil
	case map[string]interface{}:
		return rawParseMap(v, isJSON)
	case []interface{}:
		return rawParseArray(v, isJSON)
	default:
		return obj_global_null, fmt.Errorf("unable to read data into object: %+v", v)
	}
}

func rawParseMap(m map[string]interface{}, isJSON bool) (object, error) {
	ret := &objectMap{
		kvPairs: make(map[string]objectMapPair),
	}
	for k, v := range m {
		valueToAdd, err := rawParseAny(v, isJSON)
		if err != nil {
			return obj_global_null, err
		}
		kvPair := objectMapPair{
			key:   k,
			value: valueToAdd,
		}
		ret.kvPairs[k] = kvPair
	}
	return ret, nil
}

func rawParseArray(array []interface{}, isJSON bool) (object, error) {
	ret := &objectArray{
		entries: []object{},
	}
	for _, entry := range array {
		toAdd, err := rawParseAny(entry, isJSON)
		if err != nil {
			return obj_global_null, err
		}
		ret.entries = append(ret.entries, toAdd)
	}
	return ret, nil
}

func rawParseNum(num interface{}, isJSON bool) (object, error) {
	switch v := num.(type) {
	case int:
		return &objectInteger{value: int64(v)}, nil
	case int16:
		return &objectInteger{value: int64(v)}, nil
	case int32:
		return &objectInteger{value: int64(v)}, nil
	case int64:
		return &objectInteger{value: v}, nil
	case float32:
		return &objectFloat{value: float64(v)}, nil
	case float64:
		if isJSON {
			return rawParseNumJson(v), nil
		}
		return &objectFloat{value: float64(v)}, nil
	default:
		return obj_global_null, fmt.Errorf("unsupported number type: %T", v) // should only occur in custom functions
	}
}

func rawParseNumJson(num float64) object {
	if num == float64(int64(num)) {
		return &objectInteger{value: int64(num)}
	} else {
		return &objectFloat{value: num}
	}
}

func objectToInteger(obj object) int64

func objectToFloat(obj object) float64

func objectToString(obj object) string
