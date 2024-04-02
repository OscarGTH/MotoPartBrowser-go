package database

import (
	"Crawler/internal/models"
	"database/sql"
	"log"

	"github.com/spf13/viper"
)

type PSQLHandler struct {
	DB *sql.DB
}

// createDatabaseHandler connects to PostgreSQL database and returns the handler.
func CreateDatabaseHandler() *PSQLHandler {
	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", viper.GetString("database.connection_string"))
	if err != nil {
		panic(err)
	}

	// Verify the connection by pinging the database
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	log.Println("Successfully connected to PostgreSQL!")
	return &PSQLHandler{DB: db}
}

func (handler *PSQLHandler) InsertVehicle(vehicle models.Vehicle) error {
	// TODO: Optimise this.
	_, err := handler.DB.Exec("INSERT INTO Vehicles (vehicle_type, brand_name, model_name, listing_url, vehicle_id, year) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT ON CONSTRAINT unique_vehicle DO NOTHING;",
		vehicle.VehicleType, vehicle.Brand, vehicle.Model, vehicle.Url, vehicle.Identifier, vehicle.Year)
	return err
}

// InsertParts adds the parts to the database in a batch.
func (handler *PSQLHandler) InsertParts(parts []models.Part, vehicleIdentifier string) error {
	// Prepare a transaction
	tx, err := handler.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO Parts (part_name, description, part_id, vehicle_id, price, img_url, img_thumb_url) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT ON CONSTRAINT unique_part DO NOTHING;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Execute statement to add the parts.
	for _, part := range parts {
		_, err := stmt.Exec(part.Name, part.Description, part.PartIdentifier, vehicleIdentifier, part.Price, part.ImgUrl, part.ImgThumbUrl)
		if err != nil {
			return err
		}
	}
	return err
}

func (handler *PSQLHandler) GetBrands(vehicleType string) ([]string, error) {
	rows, err := handler.DB.Query("SELECT DISTINCT(brand_name) FROM Vehicles WHERE vehicle_type = $1 ORDER BY brand_name ASC;", vehicleType)
	if err != nil {
		return nil, err
	}

	var brands []string
	for rows.Next() {
		var brand string
		err = rows.Scan(&brand)
		if err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}
	return brands, nil
}

func (handler *PSQLHandler) GetModelsForBrand(vehicleType string, brandName string) ([]string, error) {
	rows, err := handler.DB.Query("SELECT DISTINCT(model_name) FROM Vehicles WHERE vehicle_type = $1 AND brand_name = $2 ORDER BY model_name ASC;", vehicleType, brandName)
	if err != nil {
		return nil, err
	}

	var models []string
	for rows.Next() {
		var model string
		err = rows.Scan(&model)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

func (handler *PSQLHandler) GetVehicle(vehicleType string, vehicleIdentifier string) (models.Vehicle, error) {
	var vehicle models.Vehicle
	err := handler.DB.QueryRow("SELECT vehicle_id, brand_name, model_name, vehicle_type, year, listing_url FROM Vehicles WHERE vehicle_type = $1 AND vehicle_id = $2 ORDER BY brand_name ASC;",
		vehicleType, vehicleIdentifier).Scan(&vehicle.Identifier,
		&vehicle.Brand, &vehicle.Model, &vehicle.VehicleType,
		&vehicle.Year, &vehicle.Url)
	if err != nil {
		return vehicle, err
	}
	return vehicle, nil
}

func (handler *PSQLHandler) GetVehicleTypes() ([]string, error) {
	rows, err := handler.DB.Query("SELECT DISTINCT(vehicle_type) FROM Vehicles ORDER BY vehicle_type ASC;")
	if err != nil {
		log.Printf("error while getting vehicle types: %v", err)
		return nil, err
	}
	defer rows.Close()

	var vehicleTypes []string
	for rows.Next() {
		var vehicleType string
		err = rows.Scan(&vehicleType)
		if err != nil {
			return nil, err
		}
		vehicleTypes = append(vehicleTypes, vehicleType)
	}
	return vehicleTypes, nil
}

func (handler *PSQLHandler) GetVehiclesForType(vehicleType string) ([]models.Vehicle, error) {
	rows, err := handler.DB.Query("SELECT vehicle_id, brand_name, model_name, vehicle_type, year, listing_url FROM Vehicles WHERE vehicle_type = $1 ORDER BY brand_name ASC;", vehicleType)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var vehicle models.Vehicle
		err = rows.Scan(&vehicle.Identifier, &vehicle.Brand, &vehicle.Model, &vehicle.VehicleType, &vehicle.Year, &vehicle.Url)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, vehicle)
	}
	return vehicles, nil
}

func (handler *PSQLHandler) GetPartsForVehicle(vehicleIdentifier string) ([]models.Part, error) {
	rows, err := handler.DB.Query("SELECT part_name, description, part_id, price, img_url, img_thumb_url FROM Parts WHERE vehicle_id = $1 ORDER BY part_name ASC;", vehicleIdentifier)
	if err != nil {
		log.Printf("error while getting parts for vehicles: %v", err)
		return nil, err
	}
	var parts []models.Part
	for rows.Next() {
		var part models.Part
		err = rows.Scan(&part.Name, &part.Description, &part.PartIdentifier, &part.Price, &part.ImgUrl, &part.ImgThumbUrl)
		if err != nil {
			log.Printf("error while scanning rows: %v", err)
		}
		parts = append(parts, part)
	}
	return parts, nil
}
