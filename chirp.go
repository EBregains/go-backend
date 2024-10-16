package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/EBregains/go-servers-learning/internal/auth"
	"github.com/EBregains/go-servers-learning/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// POST
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type response struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching parameters", err)
		return
	}

	validBody, err := ValidateAndCleanChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot validate chirp", err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validBody,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while storing chirp in database", err)
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
	})
}

// GET
func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {

	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while getting all the chirps", err)
		return
	}

	authorID := uuid.Nil
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString != "" {
		authorID, err = uuid.Parse(authorIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		if authorID != uuid.Nil && dbChirp.UserID != authorID {
			continue
		}
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

// GET {id}
func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirp
	}

	reqId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Fail at parsing uuid", err)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), reqId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
	})
}

// Validation and cleaning
func ValidateAndCleanChirp(body string) (string, error) {

	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", fmt.Errorf("the chirp exceeds maximum length: %v char", maxChirpLength)
	}
	// respond
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

// DELETE {id}
func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	// Get Chirp ID from path
	reqId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Fail at parsing uuid", err)
		return
	}
	// Get token from header and validate
	headerToken, err := auth.GetBearerToken(r.Header)
	fmt.Println(headerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting Authorization token from Header", err)
		return
	}
	userID, err := auth.ValidateJWT(headerToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't validate JWT", err)
		return
	}

	// We need to get the chirp to see its creator -- MAY GENERATE A QUERY FOR JUST GETTING THE ID --
	chirp, err := cfg.db.GetChirpById(r.Context(), reqId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find the chirp with requested ID", err)
		return
	}
	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "User doesn't have permissions to delete the chirp", err)
		return
	}

	// Delete chirp from db
	err = cfg.db.DeleteChirpById(r.Context(), reqId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
