package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"database/sql"

	rep "github.com/diskordanz/replication/rep"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Car struct {
	Number         string
	Model          string
	Year           string
	Mileage        string
	InspectionDate string
	Color          string
}

var db *sql.DB

func main() {

	dsns := "root:password@/master;"
	dsns += "root:password@/slave01;"
	dsns += "root:password@/slave02"

	db, err := rep.Open("mysql", dsns)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Some physical database is unreachable: %s", err)
	}

	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/cars", ListCars)
	http.Handle("/", r)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8080", r)

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

		var car Car
		results, err := db.Query("SELECT model FROM cars")
		if err != nil {
			log.Fatal(err)
		}
		for results.Next() {
			results.Scan(&car.Model)
			log.Printf(car.Model)
		}
	*/
}

func ListCars(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("select model,number, year, mileage, date, color from cars")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	cars := []Car{}

	for rows.Next() {
		p := Car{}
		err := rows.Scan(&p.Model, &p.Number, &p.Year, &p.Mileage, &p.InspectionDate, &p.Color)
		if err != nil {
			fmt.Println(err)
			continue
		}
		cars = append(cars, p)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cars)
}

func DeleteCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec("delete from cars_db.cars where id = ?", id)
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/", 301)
}

func GetCar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	row, err := db.Query("select * from cars where id = ?", id)
	car := Car{}
	err = row.Scan(&car.Model, &car.Number, &car.Year, &car.Mileage, &car.InspectionDate, &car.Color)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, car)

}

func EditCar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	//id := r.FormValue("id")
	//model := r.FormValue("model")
	//company := r.FormValue("company")
	//price := r.FormValue("price")

	//_, err = database.Exec("update productdb.Products set model=?, company=?, price = ? where id = ?",
	//	model, company, price, id)

	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", 301)
}

func CreateCar(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		//model := r.FormValue("model")
		//company := r.FormValue("company")
		//price := r.FormValue("price")

		//_, err = database.Exec("insert into productdb.Products (model, company, price) values (?, ?, ?)",
		//	model, company, price)

		if err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/", 301)
	} else {
		http.ServeFile(w, r, "templates/create.html")
	}
}

func Insert(car *Car) {

	stmt, err := db.Prepare(`INSERT INTO cars(
			model,
			number,			
			year,		
			mileage,		
			date,		
			color) VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		&car.Model,
		&car.Number,
		&car.Year,
		&car.Mileage,
		&car.InspectionDate,
		&car.Color)
	if err != nil {
		panic(err)
	}
	log.Println("insert successful")

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
