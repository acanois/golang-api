package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/lpernett/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type App struct {
	config *oauth2.Config
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	authConfig := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8000/auth/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	app := App{config: authConfig}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("POST /login", app.loginHandler)
	mux.HandleFunc("POST /auth", app.authHandler)
	mux.HandleFunc("POST /callback", app.oAuthCallbackHandler)

	fmt.Printf("Server listening on port: %s", port)
	http.ListenAndServe(port, mux)
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Home")
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

func (a *App) authHandler(w http.ResponseWriter, r *http.Request) {
	url := a.config.AuthCodeURL("hello world", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (a *App) oAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	t, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := a.config.Client(context.Background(), t)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()

	var v any

	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%v", v)
}
