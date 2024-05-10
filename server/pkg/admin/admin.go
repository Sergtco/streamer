package admin

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"stream/pkg/admin/views"
	"stream/pkg/database"
	"stream/pkg/structs"
	"strings"
	"sync"
)

var Cache sync.Map = sync.Map{}

func AdminIndex(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	comp := views.Index(users)
	comp.Render(r.Context(), w)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	newUser := &structs.User{}
	newUser.IsAdmin = len(r.FormValue("admin")) > 0
	newUser.Login = r.FormValue("login")
	newUser.Password = r.FormValue("password")
	newUser.Name = r.FormValue("name")
	if len(newUser.Login) != 0 && len(newUser.Password) != 0 && len(newUser.Name) != 0 {
		admin := 0
		if newUser.IsAdmin == true {
			admin = 1
		}
		newUser.Password = HashPassword(newUser.Password)
		database.InsertUser(newUser.Name, newUser.Login, newUser.Password, admin)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
	return
}

func ChangeUser(w http.ResponseWriter, r *http.Request) {
	newUser := &structs.User{}
	newUser.Login = r.FormValue("login")
	newUser.Password = r.FormValue("password")
	newUser.Name = r.FormValue("name")
	newUser.IsAdmin = len(r.FormValue("admin")) > 0
	if len(newUser.Login) != 0 && len(newUser.Password) != 0 && len(newUser.Name) != 0 {
		newUser.Password = HashPassword(newUser.Password)
		database.UpdateUser(newUser.Name, newUser.Login, newUser.Password, newUser.IsAdmin)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
	return
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	err := database.DeleteUser(login)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
	return
}

type LoginForm struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

/*
Gets json with login and hashed password.

Uses `LoginForm` struct to deserialize.
*/
func UserLogin(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer body.Close()
	var data []byte
	body.Read(data)
	form := &LoginForm{}
	err := json.Unmarshal(data, form)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := database.GetUser(form.Login)
	if user == nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokenString, err := EncodeLogin(user.Login, user.IsAdmin)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{Value: tokenString, Name: "Token"})
	w.WriteHeader(http.StatusOK)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	comp := views.Login("")
	comp.Render(r.Context(), w)
}

func CheckAdminLogin(w http.ResponseWriter, r *http.Request) {
	login, password := r.FormValue("login"), r.FormValue("password")
	user, err := database.GetUser(login)

	if user == nil || err != nil {
		comp := views.Login("User not found")
		comp.Render(r.Context(), w)
		return
	}
	if HashPassword(password) != user.Password {
		comp := views.Login("Wrong password")
		comp.Render(r.Context(), w)
		return
	}

	tokenString, err := EncodeLogin(user.Login, user.IsAdmin)

	if err != nil {
		log.Println(err)
		NotFoundHandler(w, r, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: tokenString,
	})
	Cache.Store(login, tokenString)
	if user.IsAdmin {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ValidateJwt(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := r.Cookie("token")
		if err != nil {
			log.Println(err)
			RedirectToLogin(w, r)
			return
		}
		claims, err := DecodeLogin(tokenString.Value)
		if err != nil {
			log.Println(err)
			RedirectToLogin(w, r)
			return
		}
		if exists, err := database.GetUser(claims.Login); exists != nil && err == nil {
			handler.ServeHTTP(w, r)
			return
		}
		RedirectToLogin(w, r)
		return
	})
}

func ListSongs(w http.ResponseWriter, r *http.Request) {
	songs, err := database.GetAllSongs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Oops, something wrong"))
		log.Println("Errror listing songs to admin:", err)
		return
	}
	comp := views.Songs(songs)
	err = comp.Render(r.Context(), w)
	if err != nil {
		NotFoundHandler(w, r, http.StatusInternalServerError)
		return
	}
	return
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	if UrlIsAdmin(r.URL.Path) {
		http.Redirect(w, r, "/admin/login", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusUnauthorized)
	return
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request, code int) {
	comp := views.NotFound(strconv.Itoa(code))
	w.WriteHeader(code)
	comp.Render(r.Context(), w)
}

func UrlIsAdmin(url string) bool {
	splitted := strings.Split(url, "/")
	for _, p := range splitted {
		if p == "admin" {
			return true
		}
	}
	return false
}

func HashPassword(password string) string {
	bytes := sha256.Sum256([]byte(password))
	return string(bytes[:])
}
