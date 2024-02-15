package models

type Part struct {
	Name           string
	Description    string
	PartIdentifier string
	Price          float64
	ImgUrl         string
	ImgThumbUrl    string
}

type RawPart struct {
	Name           string
	Description    string
	PartIdentifier string
	Price          string
	ImgUrl         string
	ImgThumbUrl    string
}
