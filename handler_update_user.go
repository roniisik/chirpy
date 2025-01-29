package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/roniisik/chirpy/internal/auth"
	"github.com/roniisik/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token in header", nil)
		return
	} else if strings.HasPrefix(authHeader, "Bearer ") == false {
		respondWithError(w, http.StatusBadRequest, "missing 'Bearer ' in header", nil)
		return
	}
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	user_id, err := auth.ValidateJWT(accessToken, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "access token not valid", err)
		return
	}

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type response struct {
		User
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode body", err)
		return
	}

	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to hash password", err)
		return
	}

	user, err := cfg.dbQueries.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
		Email:          params.Email,
		HashedPassword: hashedPW,
		ID:             user_id,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to query user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{User{
		ID:          user_id,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   time.Now(),
		Email:       params.Email,
		IsCherpyRed: user.IsChirpyRed,
	}})
}
