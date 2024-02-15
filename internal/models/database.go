package models

// DatabaseHandler defines the methods for interacting with the database
type DatabaseHandler interface {
	// Insert adds a new item to the database
	Insert(vehicle *Vehicle) error
}
