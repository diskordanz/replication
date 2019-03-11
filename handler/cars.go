package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/diskordanz/RestApi/api/app/model"

	"github.com/gorilla/mux"
)

func GetCars(CarService model.CarService, w http.ResponseWriter, r *http.Request) {
	cars, err := CarService.GetCars()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cars)

}

func GetCar(CarService model.CarService, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	car := model.Car{ID: id}

	if err := CarService.GetCar(&car); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, car)

}

func CreateCar(CarService model.CarService, w http.ResponseWriter, r *http.Request) {
	var car model.Car
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&car); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if err := CarService.CreateCar(&car); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, car)
}

func UpdateCar(CarService model.CarService, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	car := model.Car{ID: id}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&car); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	if err := CarService.UpdateCar(&car); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, car)
}

func DeleteCar(CarService model.CarService, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	car := model.Car{ID: id}

	if err := CarService.DeleteCar(&car); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}
