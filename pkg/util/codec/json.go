package codec

import (
	"encoding/json"
)

func EncodeToJSON(element interface{}) string {
	tempJSON, _ := json.Marshal(element)
	return string(tempJSON)
}

func ParseFromJSON(content string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	err := json.Unmarshal([]byte(content), &ret)
	return ret, err
}
