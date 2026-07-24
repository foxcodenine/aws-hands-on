package handlers

import (
	apptypes "03-dynamodb_crud_with_go/internal/app"
	"03-dynamodb_crud_with_go/internal/models"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	App *apptypes.App
}

func (h *UserHandler) Store(w http.ResponseWriter, r *http.Request) {
	// Return JSON for every response.
	w.Header().Set("Content-Type", "application/json")

	// Define the expected request body.
	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int64  `json:"age"`
	}

	var req Request

	// Decode the request body.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Check the required fields.
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" || req.Age < 0 {
		writeError(w, http.StatusBadRequest, "name, email, and a non-negative age are required")
		return
	}

	// Check whether the email is already in use.
	users, err := h.App.Repo.User.QueryByEmail(r.Context(), strings.TrimSpace(req.Email))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check email")
		return
	}
	if len(users) > 0 {
		writeError(w, http.StatusConflict, "email is already in use")
		return
	}

	// Save the user in DynamoDB.
	user, err := h.App.Repo.User.Create(r.Context(), models.CreateUserInput{
		Name:  req.Name,
		Email: strings.TrimSpace(req.Email),
		Age:   int(req.Age),
		Tags:  []string{"admin", "beta-tester"},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Return the newly created user.
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
	// Return JSON for every response.
	w.Header().Set("Content-Type", "application/json")

	// Read the user ID from the URL.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// Load the user from DynamoDB.
	user, err := h.App.Repo.User.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Return the requested user.
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Return JSON for every response.
	w.Header().Set("Content-Type", "application/json")

	// Read the user ID from the URL.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// Define and decode the fields that can be updated.
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Check the required fields.
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
		writeError(w, http.StatusBadRequest, "name and email are required")
		return
	}

	// Update the user in DynamoDB.
	user, err := h.App.Repo.User.Update(r.Context(), userID, req.Name, req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Return the updated user.
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Destroy(w http.ResponseWriter, r *http.Request) {
	// Return JSON for error responses.
	w.Header().Set("Content-Type", "application/json")

	// Read the user ID from the URL.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// Delete the user from DynamoDB.
	user, err := h.App.Repo.User.Delete(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Confirm the deletion.
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Return JSON for every response.
	w.Header().Set("Content-Type", "application/json")

	// Load all users from the table.
	users, err := h.App.Repo.User.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query users")
		return
	}

	// Return the matching users.
	_ = json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) Active(w http.ResponseWriter, r *http.Request) {
	// Return JSON for every response.
	w.Header().Set("Content-Type", "application/json")

	// Read the optional result limit.
	limit := int32(10)
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.ParseInt(rawLimit, 10, 32)
		if err != nil || parsedLimit <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive number")
			return
		}
		limit = int32(parsedLimit)
	}

	// Find only active users using the status index.
	users, err := h.App.Repo.User.QueryByStatus(r.Context(), "active", limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query active users")
		return
	}

	// Return the active users.
	_ = json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) ByEmail(w http.ResponseWriter, r *http.Request) {
	// Return JSON for every response.
	w.Header().Set("Content-Type", "application/json")

	// Read the email from the query string.
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Find users with the requested email.
	users, err := h.App.Repo.User.QueryByEmail(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find user by email")
		return
	}
	if len(users) == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if len(users) > 1 {
		writeError(w, http.StatusConflict, "more than one user has this email")
		return
	}

	// Return the matching user.
	_ = json.NewEncoder(w).Encode(users[0])
}

func writeError(w http.ResponseWriter, status int, message string) {
	// Return errors in the same JSON format.
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
