package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"forum/database"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RegisterGET(w http.ResponseWriter, r *http.Request){
	IndexHandler(w,r,"register")
}

func RegisterPOST(w http.ResponseWriter, r *http.Request){
	err:=r.ParseForm()
	if err!=nil{
		http.Error(w,"bad request",http.StatusBadRequest)
	}
	email:=r.FormValue("email")
	username:=r.FormValue("username")
	password:=r.FormValue("password")
	if email == ""||username ==""||password==""{
		http.Error(w,"missing fields",http.StatusBadRequest)
		return
	}
	//password hashing
	hash,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err!=nil{
		http.Error(w,"Server Error",http.StatusInternalServerError)
		return
	}
	_,err = database.DB.Exec(`
	INSERT INTO users(email,username,password_hash)
	VALUES (?,?,?)`,email,username,string(hash))
	if err !=nil{
		http.Error(w,"email or username already taken",http.StatusConflict)
		return
	}
	http.Redirect(w,r,"/login",http.StatusSeeOther)
}

func LoginGET(w http.ResponseWriter, r *http.Request) {
	IndexHandler(w, r, "login")
}

func LoginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	var userID int
	var hash string

	err:=database.DB.QueryRow(`
	SELECT id, password_hash FROM users WHERE username = ?`,username).Scan(&userID,&hash)
	if err == sql.ErrNoRows{
		http.Error(w,"wrong username or password",http.StatusUnauthorized)
		return
	}
	if err!=nil{
		http.Error(w,"Server Error",http.StatusInternalServerError)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))!=nil{
		http.Error(w,"invalid username or password",http.StatusUnauthorized)
		return
	}
	sessionid,err:=randomHex(32)
	if err!=nil{
		http.Error(w,"failed generating session id",http.StatusInternalServerError)
		return
	}
	expires:=time.Now().Add(24*time.Hour)

	_,err=database.DB.Exec(`
	INSERT INTO sessions(id, user_id,expires_at) VALUES(?,?,?)`,sessionid,userID,expires.UTC())
	if err !=nil{
		http.Error(w,"server error",http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:	"session_id",
		Value:	sessionid,
		Path:	"/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires: expires,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)



}

func LogoutPOST(w http.ResponseWriter, r *http.Request) {
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
	
	http.Redirect(w,r,"/",http.StatusSeeOther)
	
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}