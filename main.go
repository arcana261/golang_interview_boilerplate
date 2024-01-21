package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	DB_CONNECTION = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
)

type Article struct {
	ID          int    `json:"id" gorm:"primarykey"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func migrateDatabase() {
	dbConnection, err := sql.Open("postgres", DB_CONNECTION)
	if err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(dbConnection, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("> preparing to migrate database")
	m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	log.Printf(">> database migration done")
}

func sendOk(w http.ResponseWriter, value interface{}) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(value)
}

func main() {
	migrateDatabase()

	db, err := gorm.Open(gormPostgres.Open(DB_CONNECTION), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
	//.Queries("q", "{q:[0-9]*}") ----- required query parameter
	router.Name("getArticleById").Methods(http.MethodGet).Path("/articles/{id:[0-9]+}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		queries := r.URL.Query()
		fmt.Printf("Get Article By Id, id=%v, q=%v, queries=%v!\n", params["id"], params["q"], queries)

		var article Article
		err := db.Where("id = ?", params["id"]).First(&article).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		sendOk(w, article)
	})

	router.Name("createArticle").Methods(http.MethodPost).Path("/articles").HeadersRegexp("Content-Type", "application/json").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var newArticle Article

		err := json.NewDecoder(r.Body).Decode(&newArticle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = db.Create(&newArticle).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sendOk(w, newArticle)
	})

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("> starting server...")
	log.Fatal(srv.ListenAndServe())
}
