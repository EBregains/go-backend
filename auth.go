package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/EBregains/go-servers-learning/internal/auth"
	"github.com/EBregains/go-servers-learning/internal/database"
)

const (
	REFRESH_TOKEN_LIFESPAN time.Duration = 24 * time.Hour * 60 // equlas to 60 days
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	// Jwt Creation
	expirationTime := time.Hour
	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		expirationTime,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create acces JWT", err)
		return
	}

	// Create refresh token and store in the db
	token, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}
	dbToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  token,
		UserID: user.ID,
		ExpiresAt: sql.NullTime{
			Time: time.Now().Add(REFRESH_TOKEN_LIFESPAN),
		},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't store refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          user.ID,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        accessToken,
		RefreshToken: dbToken.Token,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	// Get refresh token from headers
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error when getting Authorization Header", err)
		return
	}

	// Get token from db and check if is expired
	dbRefreshToken, err := cfg.db.GetRefreshToken(r.Context(), headerToken)
	isExpired := dbRefreshToken.ExpiresAt.Time.Compare(time.Now())
	if err != nil || isExpired == 1 || dbRefreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Given Refresh Token does not exist or is expired", err)
		return
	}
	// If exists we need to create a new access token
	dbUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), dbRefreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user from given Refresh Token", err)
		return
	}

	token, err := auth.MakeJWT(
		dbUser.UserID,
		cfg.jwtSecret,
		time.Hour,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create acces JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from headers
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error when getting Authorization Header", err)
		return
	}
	_, err = cfg.db.RevokeToken(r.Context(), headerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error when revoking Authorization Header", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
