package utils

import (
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
