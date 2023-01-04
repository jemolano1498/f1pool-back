package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "2[$/pq(xvLEO,R)n"
	DB_NAME     = "postgres"
	DB_HOST     = "34.91.242.62"
	DB_PORT     = 5432
)

// DB set up
func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST,
		DB_PORT,
		DB_USER,
		DB_PASSWORD,
		DB_NAME)
	db, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return db
}

func connectWithConnector() (*sql.DB, error) {
	//mustGetenv := func(k string) string {
	//	v := os.Getenv(k)
	//	if v == "" {
	//		log.Fatalf("Warning: %s environment variable not set.", k)
	//	}
	//	return v
	//}
	// Note: Saving credentials in environment variables is convenient, but not
	// secure - consider a more secure solution such as
	// Cloud Secret Manager (https://cloud.google.com/secret-manager) to help
	// keep secrets safe.
	var (
		dbUser                 = DB_USER                       // e.g. 'my-db-user'
		dbPwd                  = DB_PASSWORD                   // e.g. 'my-db-password'
		dbName                 = DB_NAME                       // e.g. 'my-database'
		instanceConnectionName = "f1pools:europe-west4:f1pool" // e.g. 'project:region:instance'
		usePrivate             = "10.115.112.3"
		//dbUser                 = mustGetenv("DB_USER")                  // e.g. 'my-db-user'
		//dbPwd                  = mustGetenv("DB_PASS")                  // e.g. 'my-db-password'
		//dbName                 = mustGetenv("DB_NAME")                  // e.g. 'my-database'
		//instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME") // e.g. 'project:region:instance'
		//usePrivate             = os.Getenv("PRIVATE_IP")
	)

	d, err := cloudsqlconn.NewDialer(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %v", err)
	}
	var opts []cloudsqlconn.DialOption
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithPrivateIP())
	}
	mysql.RegisterDialContext("cloudsqlconn",
		func(ctx context.Context, addr string) (net.Conn, error) {
			return d.Dial(ctx, instanceConnectionName, opts...)
		})

	dbURI := fmt.Sprintf("%s:%s@cloudsqlconn(localhost:3306)/%s?parseTime=true",
		dbUser, dbPwd, dbName)

	dbPool, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}
	return dbPool, nil
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

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("f1pool-backend: received a request")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "f1pool-dev"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
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

// response and request handlers
func GetMovies(w http.ResponseWriter, r *http.Request) {
	//db := setupDB()
	db, err := connectWithConnector()

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

	json.NewEncoder(w).Encode(response)
}

// Create a movie

// response and request handlers
func CreateMovie(w http.ResponseWriter, r *http.Request) {
	movieID := r.FormValue("movieid")
	movieName := r.FormValue("moviename")

	var response = JsonResponse{}

	if movieID == "" || movieName == "" {
		response = JsonResponse{Type: "error", Message: "You are missing movieID or movieName parameter."}
	} else {
		db := setupDB()

		printMessage("Inserting movie into DB")

		fmt.Println("Inserting new movie with ID: " + movieID + " and name: " + movieName)

		var lastInsertID int
		err := db.QueryRow("INSERT INTO movies(movieID, movieName) VALUES($1, $2) returning id;", movieID, movieName).Scan(&lastInsertID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The movie has been inserted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

// Delete a movie

// response and request handlers
func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	movieID := params["movieid"]

	var response = JsonResponse{}

	if movieID == "" {
		response = JsonResponse{Type: "error", Message: "You are missing movieID parameter."}
	} else {
		db := setupDB()

		printMessage("Deleting movie from DB")

		_, err := db.Exec("DELETE FROM movies where movieID = $1", movieID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The movie has been deleted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

// Delete all movies

// response and request handlers
func DeleteMovies(w http.ResponseWriter, r *http.Request) {
	db := setupDB()

	printMessage("Deleting all movies...")

	_, err := db.Exec("DELETE FROM movies")

	// check errors
	checkErr(err)

	printMessage("All movies have been deleted successfully!")

	var response = JsonResponse{Type: "success", Message: "All movies have been deleted successfully!"}

	json.NewEncoder(w).Encode(response)
}
