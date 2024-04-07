package main

import (
	"Crawler/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router    *mux.Router
	DBHandler *database.PSQLHandler
}

func (a *App) Initialize() {
	a.DBHandler = database.CreateDatabaseHandler()
	a.Router = mux.NewRouter()
	a.Router.StrictSlash(true)
}

func (a *App) Run(addr string) {
	a.Router.HandleFunc("/vehicles", a.VehicleCountHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types", a.VehicleTypesHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}", a.VehiclesWithTypeHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/{vehicleId}", a.VehicleHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/{vehicleId}/parts", a.PartHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/brands", a.BrandsWithTypeHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/brands/{brandName}/models", a.ModelsForBrandHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/brands/{brandName}/models/{modelName}", a.VehiclesForModelHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/brands/{brandName}/models/{modelName}/parts", a.PartsForModelHandler).Methods("GET")
	http.Handle("/", a.Router)
	a.Router.Use(contentTypeApplicationJsonMiddleware)

	srv := &http.Server{
		Handler:      a.Router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func contentTypeApplicationJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (a *App) VehicleCountHandler(w http.ResponseWriter, r *http.Request) {
	count, err := a.DBHandler.GetVehicleCount()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(count)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) VehicleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicle, err := a.DBHandler.GetVehicle(vars["vehicleType"], vars["vehicleId"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(vehicle)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) PartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parts, err := a.DBHandler.GetPartsForVehicle(vars["vehicleId"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(parts)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) BrandsWithTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	brands, err := a.DBHandler.GetBrands(vars["vehicleType"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload, err := json.Marshal(brands)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)

}

func (a *App) ModelsForBrandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	brands, err := a.DBHandler.GetModelsForBrand(vars["vehicleType"], vars["brandName"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload, err := json.Marshal(brands)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) VehiclesWithTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicles, err := a.DBHandler.GetVehiclesForType(vars["vehicleType"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(vehicles)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) VehicleTypesHandler(w http.ResponseWriter, r *http.Request) {
	types, err := a.DBHandler.GetVehicleTypes()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(types)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) PartsForModelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parts, err := a.DBHandler.GetPartsForModel(vars["vehicleType"], vars["brandName"], vars["modelName"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(parts)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}

func (a *App) VehiclesForModelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicles, err := a.DBHandler.GetVehiclesForModel(vars["vehicleType"], vars["brandName"], vars["modelName"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(vehicles)
	if err != nil {
		log.Printf("Cannot unmarshal: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Write(payload)
}
