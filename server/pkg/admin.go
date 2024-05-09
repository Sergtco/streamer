package pkg

func AdminIndex(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie("Token"); err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else {
		// val := token.Value
		// if database.ValidateToken(token) == false {
		// 	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		// }
		fmt.Println(token)
		comp := views.Index()
		comp.Render(r.Context(), w)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	comp := views.Login()
	comp.Render(r.Context(), w)
}

func Validate(w http.ResponseWriter, r *http.Request) {
	login, password := r.FormValue("login"), r.FormValue("password")
	passSha := sha256.Sum256([]byte(password))
	var sb strings.Builder
	for i := range passSha {
		sb.WriteByte(passSha[i])
	}
	user, err := database.GetUser(login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	} else if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	} else if user.Password = sb.String() {
	}
}
