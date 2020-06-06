// Добавление нового пользователя
// curl -X POST -H 'Content-Type: application/json' -d "{\"mail\": \"asdasdasasd@mail.ru\", \"pass\": \"4545\"}" http://localhost:3333/v1/api/user/new

// Список всех пользователей
// curl -X GET http://localhost:3333/v1/api/user/all

// Авторизация по JSON данным
// curl -X GET -H 'Content-Type: application/json' -d "{\"mail\": \"that@mail.ru\", \"pass\": \"4545545\"}" http://localhost:3333/v1/api/user/auth

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	_ "github.com/lib/pq"
)

var db *sql.DB

// userD struct
type userD struct {
	ID   int
	Mail string `json:"mail"`
	Pass string `json:"pass"`
}

type users struct {
	Users []userD
}

// RoutesApp func
func RoutesApp() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/all", GetAllUsers)
	router.Get("/auth", GetAUser)
	router.Post("/new", CreateUser)
	return router
}

func initializeAPI() (*chi.Mux, *sql.DB) {
	var err error
	dbinfo := fmt.Sprintf("user=daniilak password=410552 dbname=daniilak_bd sslmode=disable")
	db, err = sql.Open("postgres", dbinfo)
	checkErr(err)

	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json ??
		middleware.Logger,          // Log API request calls
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
	)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	router.Route("/v1", func(r chi.Router) {
		r.Mount("/api/user", RoutesApp())
	})

	return router, db
}

func main() {
	router, db := initializeAPI()

	defer db.Close()

	log.Fatal(http.ListenAndServe(":3333", router))
}

// CreateUser func
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var t userD

	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &t)

	var lastInsertID int
	err := db.QueryRow("INSERT INTO users(mail,pass) VALUES($1,$2) returning did;", t.Mail, t.Pass).Scan(&lastInsertID)
	checkErr(err)

	render.JSON(w, r, render.M{"id": lastInsertID})
}

// GetAUser func
func GetAUser(w http.ResponseWriter, r *http.Request) {
	var t userD

	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &t)

	row := db.QueryRow("SELECT * FROM users WHERE mail = $1 AND pass = $2", t.Mail, t.Pass)
	var user userD
	err := row.Scan(&user.ID, &user.Mail, &user.Pass)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(404)
			render.JSON(w, r, render.M{"status": "404 not found"})
			return
		default:
			render.JSON(w, r, render.M{"status": "strange error"})
			return
		}
	}
	render.JSON(w, r, user)
	return
}

func queryAllUsers(users *users) {
	rows, err := db.Query("SELECT * FROM users")
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		user := userD{}
		err = rows.Scan(
			&user.ID,
			&user.Mail,
			&user.Pass,
		)
		checkErr(err)
		users.Users = append(users.Users, user)
	}
	err = rows.Err()
	checkErr(err)
}

// GetAllUsers func
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users := users{}

	queryAllUsers(&users)

	render.JSON(w, r, users)
}

// checkErr func
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
