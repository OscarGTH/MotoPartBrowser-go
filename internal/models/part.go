package models

type VehicleAndPart struct {
	Part    Part    `json:"part"`
	Vehicle Vehicle `json:"vehicle"`
}

type Part struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	PartIdentifier string  `json:"id"`
	Price          float64 `json:"price"`
	ImgUrl         string  `json:"img_url"`
	ImgThumbUrl    string  `json:"img_thumb_url"`
}

type RawPart struct {
	Name           string
	Description    string
	PartIdentifier string
	Price          string
	ImgUrl         string
	ImgThumbUrl    string
}
