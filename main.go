package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	rep "github.com/diskordanz/replication/rep"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Car struct {
	ID             int
	Number         string
	Model          string
	Year           string
	Mileage        string
	InspectionDate string
	Color          string
}

var db *rep.DB

func main() {

	dsns := "root:password@/master;"
	dsns += "root:password@/slave01;"
	dsns += "root:password@/slave02"

	var err error
	db, err = rep.Open("mysql", dsns)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Some physical database is unreachable: %s", err)
	}

	defer db.Close()
	/*
		err = db.Exec(`INSERT INTO cars (car_id, model,number, year, mileage, date, color)
					VALUES (1, 'dd', '333','1999', '44444','43/64', 'red'), (2, 'backy', '345245','2000', '10000','21/10', 'green');`)
		if err != nil {
			log.Fatal(err)
		}
	*/
	router := mux.NewRouter()
	router.HandleFunc("/cars/{id}", GetCar).Methods("GET")
	router.HandleFunc("/cars", ListCars).Methods("GET")
	router.HandleFunc("/cars/{id}", DeleteCar).Methods("DELETE")
	router.HandleFunc("/cars", CreateCar).Methods("POST")
	router.HandleFunc("/cars/{id}", UpdateCar).Methods("PUT")
	http.Handle("/", router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8281", router)
	/*
		err = db.Exec(`CREATE TABLE IF NOT EXISTS cars(
				car_id      integer,
				model		varchar(32),
				number		varchar(32),
				year		varchar(32),
				mileage		varchar(32),
				date		varchar(32),
				color		varchar(32));`)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Exec(`INSERT INTO cars (car_id, model,number, year, mileage, date, color)
			VALUES (1, 'dd', '333','1999', '44444','43/64', 'red'), (2, 'backy', '345245','2000', '10000','21/10', 'green');`)
		if err != nil {
			log.Fatal(err)
		}
	*/

}

func GetCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := db.Query("select * from cars where car_id = ?", id)
	if err != nil {
		log.Println(err)
	}
	defer result.Close()
	car := Car{}
	result.Next()
	err = result.Scan(&car.ID, &car.Model, &car.Number, &car.Year, &car.Mileage, &car.InspectionDate, &car.Color)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, car)
}

func ListCars(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("select * from cars")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	cars := []Car{}

	for rows.Next() {
		car := Car{}
		err := rows.Scan(&car.ID, &car.Model, &car.Number, &car.Year, &car.Mileage, &car.InspectionDate, &car.Color)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cars = append(cars, car)
	}
	respondJSON(w, http.StatusOK, cars)
}

func DeleteCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := db.Exec("delete from cars where car_id = ?", id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func UpdateCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var car Car
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&car); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if err := db.Exec("UPDATE cars SET car_id=?, model=?,number=?, year=?, mileage=?, date=?, color=? WHERE car_id=?)",
		car.ID, car.Model, car.Number, car.Year, car.Mileage, car.InspectionDate, car.Color, id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, car)
}

func CreateCar(w http.ResponseWriter, r *http.Request) {
	var car Car
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&car); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if err := db.Exec("INSERT INTO cars (car_id, model,number, year, mileage, date, color) VALUES (?,?,?,?,?,?,?)",
		car.ID, car.Model, car.Number, car.Year, car.Mileage, car.InspectionDate, car.Color); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, car)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}
