package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"forum/database"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RegisterPOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		renderPage(w, r, "error", BadRequest)
		return
	}
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	if email == "" || username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		renderPage(w, r, "error", BadRequest)
		return
	}
	//password hashing
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to hash password: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}

	_, err = database.DB.Exec(`
	INSERT INTO users(email,username,password_hash)
	VALUES (?,?,?)`, email, username, string(hash))
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		if strings.Contains(err.Error(), "users.email") {
			renderPage(w, r, "register", "Email already in use")
			return
		} else if strings.Contains(err.Error(), "users.username") {
			renderPage(w, r, "register", "Username already in use")
			return
		}
		renderPage(w, r, "error", Conflict)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to parse form: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		renderPage(w, r, "error", BadRequest)
		return
	}

	var userID int
	var hash string
	invalidAuth := "Invalid username or password"

	err := database.DB.QueryRow(`
	SELECT id, password_hash FROM users WHERE username = ?`, username).Scan(&userID, &hash)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		renderPage(w, r, "login", invalidAuth)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "database error during login: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		renderPage(w, r, "login", invalidAuth)
		return
	}
	sessionid, err := randomHex(32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to generate session ID: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}
	expires := time.Now().Add(24 * time.Hour)

	_, err = database.DB.Exec(`
	INSERT INTO sessions(id, user_id,expires_at) VALUES(?,?,?)`, sessionid, userID, expires.UTC())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteErrorLog("error.log", "failed to create session: "+err.Error())
		renderPage(w, r, "error", InternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username, // directly store username here
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true if using HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		renderPage(w, r, "error", MethodNotAllowed)
		return
	}

	c, err := r.Cookie("session_id")
	if err == nil && c.Value != "" {
		database.DB.Exec(`DELETE FROM sessions WHERE id = ?`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
