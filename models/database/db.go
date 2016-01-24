package database

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"log"
	"math/rand"
)

var db *sql.DB

func SetupDB() {
	var err error

	os.Remove("./gophertron.sqlite")
	db, err = sql.Open("sqlite3", "./gophertron.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE users (id integer not null primary key, score integer)")
	if err != nil {
		log.Fatal(err)
	}
}

func NewUser() int {
	id := rand.Int()
	for !Exists(id) {
		id = rand.Int()
	}
	db.Exec("INSERT INTO users(id, score) values($1, $2)", id, 0)

	return id
}

func Exists(id int) bool {
	rows, _ := db.Query("SELECT * FROM users where id = $1", id)
	return rows.Next()
}

func GetScore(userID int) (score int) {
	db.QueryRow("SELECT score FROM users WHERE id = $1", userID).Scan(&score)
	return
}

func IncrementScore(userID int) {
	var prevScore int
	db.QueryRow("SELECT score FROM users WHERE id = $1", userID).Scan(&prevScore)
	db.Exec("UPDATE users SET score = $1 WHERE id = $2", prevScore+1, userID)
}
