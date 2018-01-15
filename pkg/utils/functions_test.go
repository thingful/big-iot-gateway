package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertJson(t *testing.T) { // testing if it imports properly
	testcases := []struct {
		name     string
		input    string
		output   string
		offering OfferingConfig
	}{
		{
			name:   "something",
			input:  `[{"Temperature":12.11,"Humidity":15.66,"Foo":0,"Bar":"xxx"},{"Temperature":25.55,"Humidity":87,"Foo":0,"Bar":"xxx"}]`,
			output: `[{"schema:airHumidityValue":15.66,"schema:airTemperatureValue":12.11},{"schema:airHumidityValue":87,"schema:airTemperatureValue":25.55}]`,
			offering: OfferingConfig{
				Outputs: []OfferingOutput{
					OfferingOutput{
						BigiotName: "schema:airTemperatureValue",
						PipeTerm:   "Temperature",
					},
					OfferingOutput{
						BigiotName: "schema:airHumidityValue",
						PipeTerm:   "Humidity",
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			got, err := ConvertJSON([]byte(testcase.input), testcase.offering)
			if assert.Nil(t, err) {
				assert.Equal(t, testcase.output, string(got))
			}
		})
	}

}
