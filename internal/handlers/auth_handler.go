package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"go-auth-service/internal/models"
	"go-auth-service/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validation
	if req.Username == "" || req.Password == "" || req.ConfirmPassword == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.RespondWithError(w, http.StatusBadRequest, "Passwords do not match")
		return
	}

	// Check if user exists
	var existingID int
	err := h.db.QueryRow("SELECT id FROM users WHERE username = $1", req.Username).Scan(&existingID)
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "Username already exists")
		return
	} else if err != sql.ErrNoRows {
		utils.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Insert user
	var userID int
	err = h.db.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id",
		req.Username,
		string(hashedPassword),
	).Scan(&userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(req.Username, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token in database
	if err := h.storeRefreshToken(userID, refreshToken); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.AuthResponse{
		Success:      true,
		Message:      "Registration successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &models.User{
			ID:       userID,
			Username: req.Username,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validation
	if req.Username == "" || req.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get user from database
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, password_hash FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.Username, user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token in database
	if err := h.storeRefreshToken(user.ID, refreshToken); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.AuthResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &models.User{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}

func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	username, ok := r.Context().Value("username").(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.AuthResponse{
		Success: true,
		Message: "User data retrieved successfully",
		User: &models.User{
			ID:       userID,
			Username: username,
		},
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Verify and get refresh token from database
	refreshTokenHash := h.hashToken(req.RefreshToken)
	var tokenRecord models.RefreshToken
	err := h.db.QueryRow(
		"SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = $1",
		refreshTokenHash,
	).Scan(&tokenRecord.ID, &tokenRecord.UserID, &tokenRecord.TokenHash, &tokenRecord.ExpiresAt, &tokenRecord.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Check if token is expired
	if time.Now().After(tokenRecord.ExpiresAt) {
		// Delete expired token
		h.db.Exec("DELETE FROM refresh_tokens WHERE id = $1", tokenRecord.ID)
		utils.RespondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// Get user info
	var username string
	err = h.db.QueryRow("SELECT username FROM users WHERE id = $1", tokenRecord.UserID).Scan(&username)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "User not found")
		return
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := h.generateTokens(username, tokenRecord.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Delete old refresh token and store new one
	h.db.Exec("DELETE FROM refresh_tokens WHERE id = $1", tokenRecord.ID)
	if err := h.storeRefreshToken(tokenRecord.UserID, newRefreshToken); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.AuthResponse{
		Success:      true,
		Message:      "Tokens refreshed successfully",
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: &models.User{
			ID:       tokenRecord.UserID,
			Username: username,
		},
	})
}

func (h *AuthHandler) generateTokens(username string, userID int) (string, string, error) {
	// Generate Access Token (short-lived: 15 minutes)
	accessClaims := jwt.MapClaims{
		"username": username,
		"userID":   userID,
		"type":     "access",
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
		"iat":      time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token (long-lived: 7 days)
	refreshToken, err := h.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshToken, nil
}

func (h *AuthHandler) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *AuthHandler) hashToken(token string) string {
	// Hash token using SHA256
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (h *AuthHandler) storeRefreshToken(userID int, refreshToken string) error {
	refreshTokenHash := h.hashToken(refreshToken)
	expiresAt := time.Now().Add(time.Hour * 24 * 7) // 7 days

	_, err := h.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID, refreshTokenHash, expiresAt,
	)
	return err
}
