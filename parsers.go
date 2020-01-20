package onlineconf

import (
	"encoding/json"
	"strconv"
)

func (m *Module) parseSimpleParams(keyStr, valStr string) {
	m.StringParams[keyStr] = valStr
	// log.Printf("str param: %s %s", keyStr, valStr)

	if intParam, err := strconv.Atoi(valStr); err == nil {
		m.IntParams[keyStr] = intParam
		// log.Printf("int param: %s %d", keyStr, intParam)
	}
	return
}

func (m *Module) parseJSONParams(keyStr, valStr string) error {
	m.RawJSONParams[keyStr] = valStr

	byteVal := []byte(valStr)

	MapStringInterface := make(map[string]interface{})
	err := json.Unmarshal(byteVal, &MapStringInterface)
	if err != nil {
		// то скорее всего, это массив. Парсите сами!
		return nil
	}
	m.MapStringInterfaceParams[keyStr] = MapStringInterface

	mapStrStr := make(map[string]string)
	err = json.Unmarshal(byteVal, &mapStrStr)
	if err != nil {
		return nil
	}

	mapStrInt := make(map[string]int)
	mapIntStr := make(map[int]string)
	mapIntInt := make(map[int]int)

	for k, v := range mapStrStr {
		var intK, intV int
		intK, keyErr := strconv.Atoi(k)
		intV, valErr := strconv.Atoi(v)

		if keyErr == nil {
			mapIntStr[intK] = v
		}
		if valErr == nil {
			mapStrInt[k] = intV
		}
		if valErr == nil && keyErr == nil {
			mapIntInt[intK] = intV
		}
	}

	m.MapIntIntParams[keyStr] = mapIntInt
	m.MapIntStringParams[keyStr] = mapIntStr
	m.MapStringIntParams[keyStr] = mapStrInt
	m.MapStringStringParams[keyStr] = mapStrStr
	return nil
}
