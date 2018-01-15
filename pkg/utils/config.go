package utils

//OfferingConfig is xxxxx
type OfferingConfig struct {
	ID          string // id of offering, no space
	Name        string // name of offering, wiht space
	City        string // name of city
	PipeURL     string // url of thingful pipe
	Category    string // big-iot ontology represent categoty of this offering
	Datalicense string // big-iot datalicense
	Outputs     []OfferingOutput
}

// OfferingOutput is xxxxxx
type OfferingOutput struct { // this is where we match our term to their term
	BigiotName string // short name for the Output
	BigiotRDF  string // rdf of the Output
	PipeTerm   string // field name in pipe's result
}
