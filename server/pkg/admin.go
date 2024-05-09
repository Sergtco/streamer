package pkg

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"stream/pkg/database"
	"stream/pkg/views"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

var Cache sync.Map = sync.Map{}

type LoginClaims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func AdminIndex(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	comp := views.Index(users)
	comp.Render(r.Context(), w)
}

func Login(w http.ResponseWriter, r *http.Request) {
	comp := views.Login("")
	comp.Render(r.Context(), w)
}

func CheckLogin(w http.ResponseWriter, r *http.Request) {
	login, password := r.FormValue("login"), r.FormValue("password")
	user, err := database.GetUser(login)
	if user == nil || err != nil {
		comp := views.Login("User not found")
		comp.Render(r.Context(), w)
		return
	}
	passwordSha := sha256.Sum256([]byte(password))
	if string(passwordSha[:]) != user.Password {
		comp := views.Login("Wrong password")
		comp.Render(r.Context(), w)
		return
	}
	claims := &LoginClaims{Login: login}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("Secret"))
	if err != nil {
		log.Println("Error in generating JWT: ", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Token",
		Value: tokenString,
	})
	Cache.Store(login, tokenString)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func ValidateJwt(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cookie *http.Cookie
		var err error
		if cookie, err = r.Cookie("Token"); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Go authorize!"))
			return
		}
		token, err := jwt.ParseWithClaims(cookie.Value, &LoginClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Bad Signing Method!")
			}
			return []byte("Secret"), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("BadToken"))
			return
		}
		if claims, ok := token.Claims.(*LoginClaims); ok && token.Valid {
			if _, ok := Cache.Load(claims.Login); ok {
				handler.ServeHTTP(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("BadToken"))
		return
	})
}
