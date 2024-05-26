package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Race struct {
	Label     string  `json:"label"`
	Winner    *string `json:"winner"`
	DriverIds []int   `json:"driverIds"`
}

type Driver struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CarType    string `json:"carType"`
	HorsePower int    `json:"horsePower"`
	RaceId     *int   `json:"raceId"`
}

var races = make(map[int]Race)
var drivers = make(map[int]Driver)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/drivers", listDrivers)
	r.Post("/drivers", createDriver)
	r.Get("/races", listRaces)
	r.Post("/races", createRace)
	r.Get("/drivers/{id}", getDriver)
	r.Get("/races/{id}", getRace)

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}

func listDrivers(w http.ResponseWriter, r *http.Request) {
	driverList := make([]Driver, 0, len(drivers))
	for _, driver := range drivers {
		driverList = append(driverList, driver)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driverList)
}

func listRaces(w http.ResponseWriter, r *http.Request) {
	raceList := make([]Race, 0, len(races))
	for _, race := range races {
		raceList = append(raceList, race)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(raceList)
}

func getDriver(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	driver, ok := drivers[id]
	if !ok {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

func createDriver(w http.ResponseWriter, r *http.Request) {
	var driver Driver
	err := json.NewDecoder(r.Body).Decode(&driver)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if driver.ID == 0 {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	_, ok := drivers[driver.ID]
	if ok {
		http.Error(w, "Driver already exists", http.StatusConflict)
		return
	}

	drivers[driver.ID] = driver

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

func createRace(w http.ResponseWriter, r *http.Request) {
	var race Race
	err := json.NewDecoder(r.Body).Decode(&race)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	race, err = createRaceWithDrivers(drivers, race.Label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	winner := simulateRace(race, drivers)
	race.Winner = &winner

	id := len(races) + 1
	races[id] = race

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(race)
}

func getRace(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	race, ok := races[id]
	if !ok {
		http.Error(w, "Race not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(race)
}

func createRaceWithDrivers(drivers map[int]Driver, carType string) (Race, error) {
	var selectedDrivers []int
	for id, driver := range drivers {
		if driver.CarType == carType {
			if len(selectedDrivers) == 0 || abs(drivers[selectedDrivers[0]].HorsePower-driver.HorsePower) <= 50 {
				selectedDrivers = append(selectedDrivers, id)
			}
		}
	}

	if len(selectedDrivers) < 2 {
		return Race{}, errors.New("not enough drivers with the same car type and similar horsepower")
	}

	race := Race{
		Label:     carType,
		DriverIds: selectedDrivers,
	}

	return race, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func simulateRace(race Race, drivers map[int]Driver) string {
	rand.Seed(time.Now().UnixNano())
	winnerIndex := rand.Intn(len(race.DriverIds))
	winnerId := race.DriverIds[winnerIndex]
	return drivers[winnerId].Name
}
