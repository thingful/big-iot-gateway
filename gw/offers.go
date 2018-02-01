package gw

type Output struct {
	BigiotName string // short name for the Output
	BigiotRDF  string // rdf of the Output
	PipeTerm   string // field name in pipe's result
}

type Offer struct {
	ID          string  // id of offering, no space
	Name        string  // name of offering, wiht space
	City        string  // name of city
	PipeURL     string  // url of thingful pipe
	Category    string  // big-iot ontology represent categoty of this offering
	Datalicense string  // big-iot datalicense
	Price       float64 // price in cents
	Outputs     []Output
}

type OfferConf struct {
	Offers []Offer
}
