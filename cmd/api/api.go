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
	a.Router.HandleFunc("/vehicles/types", a.VehicleTypesHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}", a.VehiclesWithTypeHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/brands", a.BrandsWithTypeHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/brands/{brand}/models", a.ModelsForBrandHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/{vehicleId}", a.VehicleHandler).Methods("GET")
	a.Router.HandleFunc("/vehicles/types/{vehicleType}/{vehicleId}/parts", a.PartHandler).Methods("GET")
	http.Handle("/", a.Router)

	srv := &http.Server{
		Handler:      a.Router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func (a *App) VehicleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicle, err := a.DBHandler.GetVehicle(vars["vehicleType"], vars["vehicleId"])
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(vehicle)
	if err != nil {
		log.Println(err)
	}
	w.Write(payload)
}

func (a *App) PartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parts, err := a.DBHandler.GetPartsForVehicle(vars["vehicleId"])
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(parts)
	if err != nil {
		log.Println(err)
	}
	w.Write(payload)
}

func (a *App) BrandsWithTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	brands, err := a.DBHandler.GetBrands(vars["vehicleType"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload, err := json.Marshal(brands)
	w.Write(payload)

}

func (a *App) ModelsForBrandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	brands, err := a.DBHandler.GetModelsForBrand(vars["vehicleType"], vars["brand"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	payload, err := json.Marshal(brands)
	w.Write(payload)
}

func (a *App) VehiclesWithTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	vehicles, err := a.DBHandler.GetVehiclesForType(vars["vehicleType"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(vehicles)
	w.Write(payload)
}

func (a *App) VehicleTypesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	types, err := a.DBHandler.GetVehicleTypes()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	payload, err := json.Marshal(types)
	w.Write(payload)
}
