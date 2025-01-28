package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/roniisik/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token in header", nil)
		return
	} else if strings.HasPrefix(authHeader, "Bearer ") == false {
		respondWithError(w, http.StatusUnauthorized, "missing 'Bearer ' in header", nil)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	record, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "no matches for token in the database", err)
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "error executing query", err)
			return
		}
	}

	if record.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", err)
		return
	}

	if record.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "token has expired", nil)
		return
	}

	accessToken, err := auth.MakeJWT(record.UserID, cfg.secretKey, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, response{Token: accessToken})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "missing refresh token in header", nil)
		return
	} else if strings.HasPrefix(authHeader, "Bearer ") == false {
		respondWithError(w, http.StatusBadRequest, "missing 'Bearer ' in header", nil)
		return
	}
	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	err := cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "no matches for token in the database", err)
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "error executing query", err)
			return
		}
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
