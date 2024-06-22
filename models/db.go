package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func getConnection() {
	var err error

	if db != nil {
		return
	}

	// Init SQLite3 database
	db, err = sql.Open("sqlite3", "./app_data.db")
	if err != nil {
		log.Fatalf("🔥 failed to connect to the database: %s", err.Error())
	}

	log.Println("🚀 Connected Successfully to the Database")
}

func ReadMaterialsFromTomlFile(filePath string) ([]Material, error) {
	var data TOMLData

	if _, err := toml.DecodeFile(filePath, &data); err != nil {
		return nil, fmt.Errorf("failed to decode materials from TOML file: %w", err)
	}

	// Combine all materials into a single slice
	var materials []Material
	materials = append(materials, data.Insulation...)
	materials = append(materials, data.Other...)
	materials = append(materials, data.Wall...)

	return materials, nil
}

func MakeMigrations() {
	getConnection()

	stmt := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		username VARCHAR(64) NOT NULL
	);`

	_, err := db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}

	// Recreate the table if it already exists
	stmt = `DROP TABLE IF EXISTS materials;`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}

	stmt = `CREATE TABLE IF NOT EXISTS materials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_by INTEGER NOT NULL,
		name VARCHAR(64) NOT NULL,
		lambda REAL NOT NULL,
		price REAL NOT NULL,
		thickness REAL NOT NULL,
		description VARCHAR(255) NULL,
		FOREIGN KEY(created_by) REFERENCES users(id)
	);`

	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}

	// Add items from materials.toml file
	materials, err := ReadMaterialsFromTomlFile("./assets/data/materials.toml")
	if err != nil {
		log.Fatal(err)
	}

	stmt = `INSERT INTO materials (created_by, name, lambda, price, thickness, description)
		VALUES(?, ?, ?, ?, ?, ?);`

	for _, material := range materials {
		_, err = db.Exec(stmt, material.CreatedBy, material.Name, material.Lambda, material.Price, material.Thickness, material.Description)
		if err != nil {
			log.Fatal(err)
		}
	}

	stmt = `CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_by INTEGER NOT NULL,
		title VARCHAR(64) NOT NULL,
		description VARCHAR(255) NULL,
		status BOOLEAN DEFAULT(FALSE),
		FOREIGN KEY(created_by) REFERENCES users(id)
	);`

	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
}

/*
https://noties.io/blog/2019/08/19/sqlite-toggle-boolean/index.html
*/
