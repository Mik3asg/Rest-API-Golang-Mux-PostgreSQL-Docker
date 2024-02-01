package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID			int		`json:"id"`
	Name 		string 	`json:"name"`
	Email 		string 	`json:"email"`
	City		string  `json:"city"`
}

func main() {
	// Connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT, city TEXT)")

	if err != nil {
		log.Fatal(err)
	}

	// Create MUX router
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers(db)).Methods("GET") // get all users
	router.HandleFunc("/users/{id}", getUser(db)).Methods("GET") // get an user based on its ID
	router.HandleFunc("/users", createUser(db)).Methods("POST") // create an user 
	router.HandleFunc("/users/{id}", updateUser(db)).Methods("PUT") // update an user based on its ID
	router.HandleFunc("/users/{id}", deleteUser(db)).Methods("DELETE") // delete an user based on its ID

	// start server
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))	// add a header on the response
																				// good thing w/ GO, we can add a middleware function
}

// middleware function going back on the router to handle HTTP requests
func jsonContentTypeMiddleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}


// Function to get all users
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.City); err != nil {
				log.Fatal(err)
			}
			users = append(users, u)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(users)
	}
}


// Function to get user by id
func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var u User
		err := db.QueryRow("SELECT *FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email, &u.City)
		if err != nil {
			if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("User not found. Please ensure you have entered an existing ID."))
			return
			}
		}

		json.NewEncoder(w).Encode(u)
	}
}


// Function to create an user
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		err := db.QueryRow("INSERT INTO users (name, email, city) VALUES ($1, $2, $3) RETURNING id", u.Name, u.Email, u.City).Scan(&u.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}


// Function to update an user
func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE users SET name = $1, email = $2, city = $3 WHERE id = $4", u.Name, u.Email, u.City, id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}


// Function to delete an user
func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var u User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email, &u.City)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("User not found. Please ensure you have entered a valid ID."))
			return
		} else {
			_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
			if err != nil {
				//todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return
			}
	
			json.NewEncoder(w).Encode("User deleted.")
		}
	}
}