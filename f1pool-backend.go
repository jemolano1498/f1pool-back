package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

const (
	DbPassword = "2[$/pq(xvLEO,R)n"
	DbName     = "postgres"
	DbPort     = 5432
)

func connectWithConnector() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Warning: %s environment variable not set.", k)
		}
		return v
	}
	// Note: Saving credentials in environment variables is convenient, but not
	// secure - consider a more secure solution such as
	// Cloud Secret Manager (https://cloud.google.com/secret-manager) to help
	// keep secrets safe.
	var (
		//dbUser                 = DB_USER                       // e.g. 'my-db-user'
		dbPwd          = DbPassword                              // e.g. 'my-db-password'
		dbName         = DbName                                  // e.g. 'my-database'
		unixSocketPath = "/cloudsql/f1pools:europe-west4:f1pool" // e.g. 'project:region:instance'
		dbUser         = mustGetenv("DB_USER")                   // e.g. 'my-db-user'
		environment    = mustGetenv("ENV")
		dbHost         = mustGetenv("IP_HOST")
	)

	var dbPool = &sql.DB{}
	var err error

	if environment == "local" {
		dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbHost,
			DbPort,
			dbUser,
			DbPassword,
			dbName)
		dbPool, err = sql.Open("postgres", dbinfo)
	} else {
		dbURI := fmt.Sprintf("user=%s password=%s database=%s host=%s",
			dbUser, dbPwd, dbName, unixSocketPath)

		// dbPool is the pool of database connections.
		dbPool, err = sql.Open("pgx", dbURI)
	}

	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	return dbPool, err
}

type Movie struct {
	MovieID   string `json:"movieid"`
	MovieName string `json:"moviename"`
}

type JsonResponse struct {
	Type    string  `json:"type"`
	Data    []Movie `json:"data"`
	Message string  `json:"message"`
}

func main() {
	// Init the mux router
	router := mux.NewRouter()

	// Route handles & endpoints

	// Get all movies
	router.HandleFunc("/movies/", GetMovies).Methods("GET")

	// Create a movie
	router.HandleFunc("/movies/", CreateMovie).Methods("POST")

	// Delete a specific movie by the movieID
	router.HandleFunc("/movies/{movieid}", DeleteMovie).Methods("DELETE")

	// Delete all movies
	router.HandleFunc("/movies/", DeleteMovies).Methods("DELETE")

	// serve the app
	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Function for handling messages
func printMessage(message string) {
	fmt.Println("")
	fmt.Println(message)
	fmt.Println("")
}

// Function for handling errors
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Get all movies

// GetMovies response and request handlers
func GetMovies(w http.ResponseWriter, r *http.Request) {
	db, err := connectWithConnector()
	checkErr(err)

	printMessage("Getting movies...")

	// Get all movies from movies table that don't have movieID = "1"
	rows, err := db.Query("SELECT * FROM movies")

	// check errors
	checkErr(err)

	// var response []JsonResponse
	var movies []Movie

	// Foreach movie
	for rows.Next() {
		var id int
		var movieID string
		var movieName string

		err = rows.Scan(&id, &movieID, &movieName)

		// check errors
		checkErr(err)

		movies = append(movies, Movie{MovieID: movieID, MovieName: movieName})
	}

	var response = JsonResponse{Type: "success", Data: movies}

	err = json.NewEncoder(w).Encode(response)
	checkErr(err)
}

// Create a movie

// CreateMovie response and request handlers
func CreateMovie(w http.ResponseWriter, r *http.Request) {
	movieID := r.FormValue("movieid")
	movieName := r.FormValue("moviename")

	var response = JsonResponse{}

	if movieID == "" || movieName == "" {
		response = JsonResponse{Type: "error", Message: "You are missing movieID or movieName parameter."}
	} else {
		db, err := connectWithConnector()
		checkErr(err)

		printMessage("Inserting movie into DB")

		fmt.Println("Inserting new movie with ID: " + movieID + " and name: " + movieName)

		var lastInsertID int
		err = db.QueryRow("INSERT INTO movies(movieID, movieName) VALUES($1, $2) returning id;", movieID, movieName).Scan(&lastInsertID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The movie has been inserted successfully!"}
	}

	err := json.NewEncoder(w).Encode(response)
	checkErr(err)
}

// Delete a movie

// DeleteMovie response and request handlers
func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	movieID := params["movieid"]

	var response = JsonResponse{}

	if movieID == "" {
		response = JsonResponse{Type: "error", Message: "You are missing movieID parameter."}
	} else {
		db, err := connectWithConnector()
		checkErr(err)

		printMessage("Deleting movie from DB")

		_, err = db.Exec("DELETE FROM movies where movieID = $1", movieID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The movie has been deleted successfully!"}
	}

	err := json.NewEncoder(w).Encode(response)
	checkErr(err)
}

// Delete all movies

// DeleteMovies response and request handlers
func DeleteMovies(w http.ResponseWriter, r *http.Request) {
	db, err := connectWithConnector()
	checkErr(err)

	printMessage("Deleting all movies...")

	_, err = db.Exec("DELETE FROM movies")

	// check errors
	checkErr(err)

	printMessage("All movies have been deleted successfully!")

	var response = JsonResponse{Type: "success", Message: "All movies have been deleted successfully!"}

	err = json.NewEncoder(w).Encode(response)
	checkErr(err)
}
