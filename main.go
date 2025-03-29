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

// Match representa un partido con todos los campos.
type Match struct {
	ID          int    `json:"id"`
	HomeTeam    string `json:"homeTeam"`
	AwayTeam    string `json:"awayTeam"`
	MatchDate   string `json:"matchDate"`
	HomeGoals   int    `json:"homeGoals"`
	AwayGoals   int    `json:"awayGoals"`
	YellowCards int    `json:"yellowCards"`
	RedCards    int    `json:"redCards"`
	ExtraTime   int    `json:"extraTime"`
}

var db *sql.DB

func main() {
	var err error
	// Variables de entorno para la conexión a la base de datos.
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Ajusta al puerto correcto (por ejemplo, "5436" si es el caso)
	}

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

	// Configuración del router con Gorilla Mux.
	router := mux.NewRouter()

	// Endpoints obligatorios
	router.HandleFunc("/api/matches", getMatches).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/matches/{id}", getMatch).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/matches", createMatch).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE", "OPTIONS")

	// Endpoints PATCH para actualizar goles, tarjetas y tiempo extra.
	router.HandleFunc("/api/matches/{id}/goals", updateGoals).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/matches/{id}/yellowcards", updateYellowCards).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/matches/{id}/redcards", updateRedCards).Methods("PATCH", "OPTIONS")
	router.HandleFunc("/api/matches/{id}/extratime", updateExtraTime).Methods("PATCH", "OPTIONS")

	// Aplicar middleware CORS
	handler := corsMiddleware(router)

	// Levantar el servidor en el puerto 8080
	log.Println("Servidor corriendo en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Middleware que configura CORS.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permite peticiones desde cualquier origen.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getMatches retorna todos los partidos con todos los campos.
func getMatches(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, home_team, away_team, match_date, 
		       home_goals, away_goals, yellowcards, redcards, extratime
		FROM matches
	`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		err := rows.Scan(
			&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate,
			&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// getMatch retorna un partido por su ID.
func getMatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	query := `
		SELECT id, home_team, away_team, match_date, 
		       home_goals, away_goals, yellowcards, redcards, extratime
		FROM matches WHERE id = $1
	`
	var m Match
	err = db.QueryRow(query, id).Scan(
		&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate,
		&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime,
	)
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

// createMatch inserta un nuevo partido. Las columnas adicionales se inicializan en 0.
func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `
		INSERT INTO matches (home_team, away_team, match_date, home_goals, away_goals, yellowcards, redcards, extratime)
		VALUES ($1, $2, $3, 0, 0, 0, 0, 0) RETURNING id
	`
	err := db.QueryRow(query, m.HomeTeam, m.AwayTeam, m.MatchDate).Scan(&m.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Se retornan los datos del partido con los valores por defecto para las columnas adicionales.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// updateMatch actualiza los campos básicos de un partido.
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
	query := `
		UPDATE matches 
		SET home_team = $1, away_team = $2, match_date = $3 
		WHERE id = $4
	`
	result, err := db.Exec(query, m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}
	m.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// deleteMatch elimina un partido por ID.
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
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateGoals actualiza los goles de un partido.
// Se espera un JSON con "homeGoals" y "awayGoals".
func updateGoals(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var payload struct {
		HomeGoals int `json:"homeGoals"`
		AwayGoals int `json:"awayGoals"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := "UPDATE matches SET home_goals = $1, away_goals = $2 WHERE id = $3"
	result, err := db.Exec(query, payload.HomeGoals, payload.AwayGoals, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateYellowCards registra tarjetas amarillas en un partido.
// Se espera un JSON con "yellowCards".
func updateYellowCards(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var payload struct {
		YellowCards int `json:"yellowCards"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := "UPDATE matches SET yellowcards = $1 WHERE id = $2"
	result, err := db.Exec(query, payload.YellowCards, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateRedCards registra tarjetas rojas en un partido.
// Se espera un JSON con "redCards".
func updateRedCards(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var payload struct {
		RedCards int `json:"redCards"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := "UPDATE matches SET redcards = $1 WHERE id = $2"
	result, err := db.Exec(query, payload.RedCards, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateExtraTime registra el tiempo extra en un partido.
// Se espera un JSON con "extraTime".
func updateExtraTime(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var payload struct {
		ExtraTime int `json:"extraTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := "UPDATE matches SET extratime = $1 WHERE id = $2"
	result, err := db.Exec(query, payload.ExtraTime, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
