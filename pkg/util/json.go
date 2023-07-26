package util

import (
	"encoding/json"
)

func EncodeToJSON(element interface{}) string {
	tempJSON, _ := json.Marshal(element)
	return string(tempJSON)
}

func ParseFromJSON() string {
	tempJSON, _ := json.Marshal(element)
	return string(tempJSON)
}
