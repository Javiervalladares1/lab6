package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
}

var db *sql.DB

func main() {
	var err error
	// Variables de entorno para la conexión a la base de datos
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Cambia al puerto que necesites (por ejemplo, "5436" si es tu caso)
	}

	// Cadena de conexión a PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error conectando a la base de datos: ", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("No se pudo alcanzar la base de datos: ", err)
	}
	fmt.Println("Conectado a la base de datos correctamente.")

	// Configuración del router
	router := mux.NewRouter()

	// Endpoints obligatorios (incluyendo soporte para OPTIONS)
	router.HandleFunc("/api/matches", getMatches).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/matches/{id}", getMatch).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/matches", createMatch).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE", "OPTIONS")

	// Envuelve el router con el middleware CORS
	handler := corsMiddleware(router)

	// Levantar el servidor en el puerto 8080
	log.Println("Servidor corriendo en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Middleware CORS: agrega los encabezados para permitir peticiones desde otros orígenes
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permite peticiones desde cualquier origen
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Métodos permitidos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Encabezados permitidos
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Si la petición es preflight (OPTIONS), responder inmediatamente
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, home_team, away_team, match_date FROM matches")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func getMatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var m Match
	err = db.QueryRow("SELECT id, home_team, away_team, match_date FROM matches WHERE id = $1", id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Partido no encontrado", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var id int
	err := db.QueryRow("INSERT INTO matches (home_team, away_team, match_date) VALUES ($1, $2, $3) RETURNING id",
		m.HomeTeam, m.AwayTeam, m.MatchDate).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.ID = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func updateMatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var m Match
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE matches SET home_team = $1, away_team = $2, match_date = $3 WHERE id = $4",
		m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	m.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func deleteMatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("DELETE FROM matches WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
