package models

type Vehicle struct {
	Name        string `json:"-"`
	Brand       string `json:"Brand"`
	Model       string `json:"Model"`
	VehicleType string `json:"VehicleType"`
	Identifier  string `json:"Identifier"`
	Year        int    `json:"Year"`
	Url         string `json:"Url"`
	Parts       []Part `json:"Parts"`
}

type RawVehicle struct {
	Name     string
	Brand    string
	Model    string
	Year     string
	Url      string
	RawParts []RawPart
}
