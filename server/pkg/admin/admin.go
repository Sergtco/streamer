package admin

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"stream/pkg/admin/views"
	"stream/pkg/database"
	"stream/pkg/structs"
	"strings"
	"sync"
	"time"
)

var Cache sync.Map = sync.Map{}

// swagger:response htmlPage
type HtmlPage struct {
	Data string `json:"data"`
}

// swagger:route GET /admin admin adminIndex
//
// Admin index page
// responses:
//
//	200: htmlPage
//	400: badRequest
//	500: internalServerError
func AdminIndex(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	comp := views.Index(users)
	comp.Render(r.Context(), w)
}

// swagger:route POST /admin/add_user admin user addUser
//
// Add user
// responses:
//
//	303: seeOther
//	400: badRequest
//	500: internalServerError
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

// swagger:route POST /admin/change_user admin user changeUser
//
// Change user's data.
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
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

// swagger:route DELETE /admin/delete_user admin user deleteUser
//
// Completely delete user.
// responses:
//
//	303: seeOther
//	400: badRequest
//	500: internalServerError
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	err := database.DeleteUser(login)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
	return
}

//swagger:parameters userLogin
type LoginForm struct {
	//Login
	//in:body
	Login string `json:"login"`
	//Password
	//in:body
	Password string `json:"password"`
}

type LoginResponse struct {
	//Token
	//in:Header
	Token string `json:"token"`
}

// swagger:route POST /login user userLogin
//
// Gets json with login and hashed password.
//
// Uses `LoginForm` struct to deserialize.
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
func UserLogin(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	form := &LoginForm{}
	err := json.Unmarshal(data, form)
	if err != nil {
		log.Println(err)
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
	exp := time.Now().Add(time.Hour * 24 * 365)
	http.SetCookie(w, &http.Cookie{
		Value:   tokenString,
		Name:    "token",
		Expires: exp,
		MaxAge:  math.MaxInt64,
		Path:    "/",
	})
	w.WriteHeader(http.StatusOK)
}

// swagger:route GET /admin/login admin adminLogin
//
// Gets from with login and hashed password.
//
// Uses `LoginForm` struct to deserialize.
// responses:
//
//	200: statusOk
//	400: badRequest
//	500: internalServerError
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

	exp := time.Now().Add(time.Hour)
	http.SetCookie(w, &http.Cookie{
		Value:   tokenString,
		Name:    "token",
		Expires: exp,
		MaxAge:  900000,
		Path:    "/",
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

// swagger:route GET /admin/songs admin adminSongs
//
// Shows list of all songs.
// responses:
//
//	200: htmlPage
//	400: badRequest
//	500: internalServerError
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
