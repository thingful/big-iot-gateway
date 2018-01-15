package utils

import (
	"encoding/json"
	"strings"
)

func GetOfferingIndex(id string, offerings []OfferingConfig) int {
	offeringIndex := -1
	for i, offering := range offerings {
		if id == strings.ToLower(offering.ID) {
			offeringIndex = i
			break
		}
	}
	return offeringIndex
}

// ConvertJSON takes pipe json and change to big-iot json depends on offerinConfig provide
func ConvertJSON(pipeJson []byte, offering OfferingConfig) ([]byte, error) {

	output := []map[string]interface{}{}

	var m interface{}
	err := json.Unmarshal(pipeJson, &m)
	if err != nil {
		return nil, err
	}

	j := m.([]interface{}) //type case to slice first

	for _, member := range j {
		pipeData := member.(map[string]interface{}) // then for each member, cast to map string interface

		bigiotData := map[string]interface{}{} // make temporary var

		for _, output := range offering.Outputs {
			if val, ok := pipeData[output.PipeTerm]; ok { // find if the key exist, if it does assign it
				bigiotData[output.BigiotName] = val
			} else { // if it doesn't exist, assing default value
				bigiotData[output.BigiotName] = ""
			}
		}

		output = append(output, bigiotData)
	}

	s, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	return s, nil
}
