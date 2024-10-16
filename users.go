package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/EBregains/go-servers-learning/internal/auth"
	"github.com/EBregains/go-servers-learning/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
	}
	// decode req
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldnt decode parameters", err)
		return
	}

	// Hash the password to store it
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while hashing password", err)
		return
	}

	// do the ting with req
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
	}
	// Get params
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while decodind params", err)
		return
	}
	// Get Token And validate JWT
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting Acces Token form header", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "JWT is not valid", err)
		return
	}
	// Hash password an Update User in db
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while hashing user password", err)
		return
	}

	dbUser, err := cfg.db.UpdateUserPassAndEmail(r.Context(), database.UpdateUserPassAndEmailParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while updating users email and hashed password", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpyRed: dbUser.IsChirpyRed,
		},
	})

}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserID string `json:"user_id"`
	}
	type parameters struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}
	// get API sign from headers
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get Api Key", err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Provided ApiKey doesn't match", err)
	}

	// decode
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while decoding params", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userUUID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing user_id to UUID", err)
		return
	}

	_, err = cfg.db.GetUserByID(r.Context(), userUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	// Red is the premium plan of chirp
	_, err = cfg.db.UpdateUserToRed(r.Context(), userUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
