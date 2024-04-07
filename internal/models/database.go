package models

// DatabaseHandler defines the methods for interacting with the database
type DatabaseHandler interface {
	InsertVehicles(vehicles []Vehicle) error
	InsertParts(vehicles []Vehicle) error
	GetVehicleCount() (int, error)
	GetVehicleTypes() ([]string, error)
	GetVehiclesForType(vehicleType string) ([]Vehicle, error)
	GetBrands(vehicleType string) ([]string, error)
	GetModelsForBrand(vehicleType string, brandName string) ([]string, error)
	GetVehiclesForModel(vehicleType string, brandName string, modelName string) ([]Vehicle, error)
	GetVehicle(vehicleType string, vehicleIdentifier string) (Vehicle, error)
	GetPartsForVehicle(vehicleIdentifier string) (Vehicle, error)
	GetPartsForModel(vehicleType string, brandName string, modelName string) ([]Vehicle, error)
}
