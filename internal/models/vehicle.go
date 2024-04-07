package models

type Vehicle struct {
	Name        string `json:"-"`
	Brand       string `json:"brand_name"`
	Model       string `json:"model_name"`
	VehicleType string `json:"vehicleType"`
	Identifier  string `json:"vehicleId"`
	Year        int    `json:"year"`
	Url         string `json:"url"`
	Parts       []Part `json:"parts"`
}

type RawVehicle struct {
	Name     string
	Brand    string
	Model    string
	Year     string
	Url      string
	RawParts []RawPart
}
