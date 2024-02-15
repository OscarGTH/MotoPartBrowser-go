package models

type Vehicle struct {
	Name        string
	Brand       string
	Model       string
	VehicleType string
	Year        int
	Url         string
	Parts       []Part
}

type RawVehicle struct {
	Name     string
	Brand    string
	Model    string
	Year     string
	Url      string
	RawParts []RawPart
}
