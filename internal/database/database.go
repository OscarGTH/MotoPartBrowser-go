package database

import (
	"Crawler/internal/models"
	"database/sql"
	"errors"
	"log"

	"github.com/schollz/progressbar/v3"
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

func hasDuplicateVehicleIDs(vehicles []models.Vehicle) bool {
	// Create a map to store seen vehicle identifiers
	seen := make(map[string]bool)

	// Iterate over the vehicles
	for _, vehicle := range vehicles {
		// Check if the current vehicle identifier has been seen before
		if seen[vehicle.Identifier] {
			log.Println(vehicle.Name)
			return true // Duplicate vehicle identifier found
		}
		// Mark the current vehicle identifier as seen
		seen[vehicle.Identifier] = true
	}

	// No duplicate vehicle identifiers found
	return false
}

func (handler *PSQLHandler) InsertVehicles(vehicles []models.Vehicle) error {
	duplicates := hasDuplicateVehicleIDs(vehicles)
	if duplicates {
		return errors.New("duplicate id found")
	}
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
			err = tx.Commit()
			if err != nil {
				log.Printf("error while committing transaction: %v", err)
			}
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO Vehicles (vehicle_type, brand_name, model_name, listing_url, vehicle_id, year) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	bar := progressbar.Default(int64(len(vehicles)))
	for _, vehicle := range vehicles {
		_, err := stmt.Exec(vehicle.VehicleType, vehicle.Brand, vehicle.Model, vehicle.Url, vehicle.Identifier, vehicle.Year)
		if err != nil {
			return err
		}
		bar.Add(1)
	}
	return nil
}

// InsertParts adds the parts to the database in a batch.
func (handler *PSQLHandler) InsertParts(vehicles []models.Vehicle) error {
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
			err = tx.Commit()
			if err != nil {
				log.Printf("error while committing transaction: %v", err)
			}
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO Parts (part_name, description, part_id, vehicle_id, price, img_url, img_thumb_url) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	var totalPartCount int
	// Get total part count for progress bar.
	for _, vehicle := range vehicles {
		totalPartCount += len(vehicle.Parts)
	}
	bar := progressbar.Default(int64(totalPartCount))

	for _, vehicle := range vehicles {
		vehicleId := vehicle.Identifier
		for _, part := range vehicle.Parts {
			_, err := stmt.Exec(part.Name, part.Description, part.PartIdentifier, vehicleId, part.Price, part.ImgUrl, part.ImgThumbUrl)
			if err != nil {
				return err
			}
			bar.Add(1)
		}
	}
	return nil
}

func (handler *PSQLHandler) GetVehicleCount() (int, error) {
	count := 0
	err := handler.DB.QueryRow("SELECT COUNT(vehicle_id) FROM Vehicles;").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
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

func (handler *PSQLHandler) GetVehiclesForModel(vehicleType string, brandName string, modelName string) ([]models.Vehicle, error) {
	rows, err := handler.DB.Query("SELECT vehicle_id, brand_name, model_name, vehicle_type, year, listing_url FROM Vehicles WHERE vehicle_type = $1 AND brand_name = $2 AND model_name = $3 ORDER BY year ASC;", vehicleType, brandName, modelName)
	if err != nil {
		log.Printf("error while getting vehicles for model: %v", err)
		return nil, err
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

func (handler *PSQLHandler) GetPartsForVehicle(vehicleIdentifier string) (models.Vehicle, error) {
	rows, err := handler.DB.Query("SELECT V.vehicle_id, V.year, V.model_name, V.brand_name, P.part_name, P.description, P.part_id, P.price, P.img_url, P.img_thumb_url FROM Vehicles V INNER JOIN Parts P ON V.vehicle_id = P.vehicle_id WHERE V.vehicle_id = $1 ORDER BY P.part_name ASC;", vehicleIdentifier)
	var vehicle models.Vehicle
	if err != nil {
		log.Printf("error while getting parts for a vehicle: %v", err)
		return vehicle, err
	}
	defer rows.Close()

	vehicle.Parts = []models.Part{}
	for rows.Next() {
		var part models.Part
		err := rows.Scan(
			&vehicle.Identifier,
			&vehicle.Year,
			&vehicle.Model,
			&vehicle.Brand,
			&part.Name,
			&part.Description,
			&part.PartIdentifier,
			&part.Price,
			&part.ImgUrl,
			&part.ImgThumbUrl,
		)
		if err != nil {
			log.Printf("error while scanning database response for parts for a vehicle: %v", err)
			return models.Vehicle{}, err
		}
		vehicle.Parts = append(vehicle.Parts, part)
	}
	return vehicle, nil
}

func (handler *PSQLHandler) GetPartsForModel(vehicleType string, brandName string, modelName string) ([]models.Vehicle, error) {
	rows, err := handler.DB.Query("SELECT V.vehicle_id, V.year, V.model_name, V.brand_name, P.part_name, P.description, P.part_id, P.price, P.img_url, P.img_thumb_url FROM Vehicles V INNER JOIN Parts P ON V.vehicle_id = P.vehicle_id WHERE V.vehicle_type = $1 AND V.brand_name = $2 AND V.model_name = $3 ORDER BY V.year ASC;", vehicleType, brandName, modelName)
	if err != nil {
		log.Printf("error while getting parts for model: %v", err)
		return nil, err
	}
	defer rows.Close()

	vehicleMap := make(map[string]models.Vehicle)
	for rows.Next() {
		var vehicleID string
		var vehicle models.Vehicle
		var part models.Part
		err = rows.Scan(&vehicleID,
			&vehicle.Year,
			&vehicle.Model,
			&vehicle.Brand,
			&part.Name,
			&part.Description,
			&part.PartIdentifier,
			&part.Price,
			&part.ImgUrl,
			&part.ImgThumbUrl,
		)
		if err != nil {
			log.Printf("error while scanning database response for parts for model: %v", err)
			return nil, err
		}

		// Check if a vehicle with the current vehicleID already exists in the vehicleMap
		existingVehicle, exists := vehicleMap[vehicleID]
		if exists {
			// Append the part to the existing vehicle's Parts slice
			existingVehicle.Parts = append(existingVehicle.Parts, part)
			// Update the vehicleMap with the modified vehicle
			vehicleMap[vehicleID] = existingVehicle
		} else {
			// Create a new vehicle object
			vehicle.Identifier = vehicleID
			// Initialize the Parts slice of the new vehicle with the current part
			vehicle.Parts = []models.Part{part}
			// Add the new vehicle to the vehicleMap
			vehicleMap[vehicleID] = vehicle
		}

	}

	// Convert the map to a slice
	vehicles := make([]models.Vehicle, 0, len(vehicleMap))
	for _, value := range vehicleMap {
		vehicles = append(vehicles, value)
	}

	return vehicles, nil
}
