package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	rethink "gopkg.in/gorethink/gorethink.v3"
)

type App struct {
	Router         *mux.Router
	RethinkSession *rethink.Session
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) Initialize(address, username, password, database string) {
	var err error

	a.RethinkSession, err = rethink.Connect(rethink.ConnectOpts{
		Address:  address,
		Database: database,
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	a.Router = mux.NewRouter()
	a.InitializeRoutes()
	a.InitializeDb(a.RethinkSession, database)
}

func (a *App) InitializeDb(session *rethink.Session, database string) {
	_, err := rethink.DBCreate(database).RunWrite(session)
	if err != nil {
		log.Print(err.Error())
	}

	_, err = rethink.TableCreate("users").RunWrite(session)
	if err != nil {
		log.Print(err.Error())
	}

	_, err = rethink.TableCreate("questions").RunWrite(session)
	if err != nil {
		log.Print(err.Error())
	}
}

func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/api/users", a.getUsers).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/api/users", a.createUser).Methods("POST")

	a.Router.HandleFunc("/api/users/{id}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/api/users/{id}/score", a.getScore).Methods("GET")

	a.Router.HandleFunc("/api/questions", a.getQuestions).Methods("GET")
	a.Router.HandleFunc("/api/questions/{id}", a.validateQuestion).Methods("POST")

}

func (a *App) Run(addr string) {
	origins := handlers.AllowedOrigins([]string{"*"})
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "content-type"})
	methods := handlers.AllowedMethods([]string{"GET", "DELETE", "HEAD", "POST", "PUT", "OPTIONS"})
	log.Fatal(http.ListenAndServe(":8888", handlers.CORS(headers, origins, methods)(a.Router)))
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	cursor, err := rethink.Table("users").Run(a.RethinkSession)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	users, err := getUsers(cursor)

	respondWithJSON(w, http.StatusOK, users)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	// TODO: test the receives values
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	hash, err := MD5FromString(u.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	u.ID = hash

	if err := u.createUser(a.RethinkSession); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	u := User{ID: vars["id"]}
	if err := u.getUser(a.RethinkSession); err != nil {
		switch err {
		case rethink.ErrEmptyResult:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) getScore(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, nil)
}

func (a *App) getQuestions(w http.ResponseWriter, r *http.Request) {
	cursor, err := rethink.Table("questions").Run(a.RethinkSession)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	questions, err := getQuestions(cursor)

	respondWithJSON(w, http.StatusOK, questions)
}

func (a *App) validateQuestion(w http.ResponseWriter, r *http.Request) {
	var v validateQuestion
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&v); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// retrieve question
	vars := mux.Vars(r)
	questionID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "invalid question id")
		return
	}
	q := Question{ID: questionID}

	if err := q.getQuestion(a.RethinkSession); err != nil {
		switch err {
		case rethink.ErrEmptyResult:
			respondWithError(w, http.StatusNotFound, "Question not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// retrieve user
	u := User{ID: v.UserID}
	if err := u.getUser(a.RethinkSession); err != nil {
		switch err {
		case rethink.ErrEmptyResult:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if q.Index == v.Answer {
		u.hitScore(a.RethinkSession)
	}

	respondWithJSON(w, http.StatusOK, u)
}
