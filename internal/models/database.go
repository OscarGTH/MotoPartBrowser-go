package models

// DatabaseHandler defines the methods for interacting with the database
type DatabaseHandler interface {
	InsertVehicle(vehicles Vehicle) error
	InsertPart(parts []Part, vehicleId string) error
	GetVehicleTypes() ([]string, error)
	GetVehiclesForType(vehicleType string) ([]Vehicle, error)
	GetBrands(vehicleType string) ([]string, error)
	GetModelsForBrand(vehicleType string, brandName string) ([]string, error)
	GetVehicle(vehicleType string, vehicleIdentifier string) (Vehicle, error)
	GetPartsForVehicle(vehicleIdentifier string) ([]Part, error)
}
