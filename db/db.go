package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"pinterestdownload/config"
	"slices"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID    int
	LOGIN string
	//PASSWORD string
}

type Boards struct {
	ID         int
	USER_ID    int
	BOARD_NAME string
	BOARD_URL  string
}

type Pins struct {
	ID       int
	BOARD_ID int
	PIN_URL  string
}

// ConnectDB Connects to database.
func ConnectDB() (*sql.DB, error) {
	databasePath := filepath.Join(config.GetConfigPath(), "Pins", "DB.db")
	PinDB, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, err
	}
	return PinDB, nil
}

func CheckUserDB(user string) int {

	var userID int
	DBPath := filepath.Join(config.GetConfigPath(), "Pins", "DB.db")
	_, err := os.Stat(DBPath)
	if err != nil {
		fmt.Println("\nDatabase not exist. Creating database")
		NewDBCreate(user)
		userID = GetUserID(user)
	}

	if GetUserID(user) == 0 {
		fmt.Println("\nUser not found in database. Creating user")
		NewUserCreate(user)
		userID = GetUserID(user)
	} else {
		userID = GetUserID(user)
	}

	return userID
}

// NewDBCreate Create new database and add first user.
func NewDBCreate(userLogin string) {
	newpath := filepath.Join(config.GetConfigPath(), "Pins")
	file := newpath + "/DB.db"
	os.Create(file)
	PinDB, err := sql.Open("sqlite3", file)
	if err != nil {
		fmt.Println(err)
	}
	createTables := `
				CREATE TABLE IF NOT EXISTS USERS (
					ID INTEGER NOT NULL PRIMARY KEY,
					LOGIN TEXT NOT NULL
				);

				CREATE TABLE IF NOT EXISTS BOARDS (
					ID INTEGER NOT NULL PRIMARY KEY,
					USER_ID INTEGER NOT NULL,
					BOARD_NAME TEXT NOT NULL,
					BOARD_URL TEXT NOT NULL
				);

				CREATE TABLE IF NOT EXISTS PINS (
					ID INTEGER NOT NULL PRIMARY KEY,
					BOARD_ID TEXT NOT NULL,
					PIN_URL TEXT NOT NULL
				);
				`

	if _, err := PinDB.Exec(createTables); err != nil {
		fmt.Println(err)
	}

	if _, err := PinDB.Exec("INSERT INTO USERS VALUES(NULL,?);", userLogin); err != nil {
		fmt.Println(err)
	}

	PinDB.Close()
}

func NewUserCreate(userLogin string) {

	newpath := filepath.Join(config.GetConfigPath(), "Pins")
	file := newpath + "/DB.db"
	PinDB, err := sql.Open("sqlite3", file)

	if err != nil {
		fmt.Println(err)
	}

	if _, err := PinDB.Exec("INSERT INTO USERS VALUES(NULL,?);", userLogin); err != nil {
		fmt.Println(err)
	}

	PinDB.Close()
}

func GetUserID(userLogin string) int {

	PinDB, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}

	var userID int
	err = PinDB.QueryRow("SELECT ID FROM USERS WHERE LOGIN = ?;", userLogin).Scan(&userID)
	if err != nil {
		fmt.Println(err)
	}

	return userID
}

// Add information about boards in database
func DBAddBoards(userID int, boardNames []string, boardLinks []string) {

	PinDB, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}

	stmt, err := PinDB.Prepare("INSERT INTO BOARDS(ID, USER_ID, BOARD_NAME, BOARD_URL) VALUES( ?, ?, ?, ? )")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	userBoards, err := PinDB.Query("SELECT BOARD_NAME FROM BOARDS WHERE USER_ID = ?;", userID)
	if err != nil {
		fmt.Println(err)
	}
	defer userBoards.Close()

	var boardsDB []string
	var tempName string

	for userBoards.Next() {
		userBoards.Scan(&tempName)
		boardsDB = append(boardsDB, tempName)
	}

	for i, v := range boardNames {
		if !slices.Contains(boardsDB, v) {
			if _, err := stmt.Exec(nil, userID, v, boardLinks[i]); err != nil {
				log.Fatal(err)
			}
		}
	}
	PinDB.Close()
}

func DBAddPins(userID int, boardName string, pinLinks []string) {

	PinDB, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}

	var boardID int
	err = PinDB.QueryRow("SELECT ID FROM BOARDS WHERE USER_ID = ? AND BOARD_NAME = ?;", userID, boardName).Scan(&boardID)
	if err != nil {
		fmt.Println(err)
	}

	stmt, err := PinDB.Prepare("INSERT INTO PINS(ID, BOARD_ID, PIN_URL) VALUES( ?, ?, ? )")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	userPins, err := PinDB.Query("SELECT PIN_URL FROM PINS WHERE BOARD_ID = ?;", boardID)
	if err != nil {
		fmt.Println(err)
	}
	defer userPins.Close()

	var pinsDB []string
	var tempName string

	for userPins.Next() {
		userPins.Scan(&tempName)
		pinsDB = append(pinsDB, tempName)
	}
	// compare values from database, find already exist
	for _, v := range pinLinks {
		if !slices.Contains(pinsDB, v) {
			if _, err := stmt.Exec(nil, boardID, v); err != nil {
				log.Fatal(err)
			}
		}
	}
	PinDB.Close()
}
