package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pelletier/go-toml/v2"
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

func addMaterialsFromTomlFile(path string) error {
	// Read the TOML file
	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	// Define a struct to hold the TOML data
	type Material struct {
		CreatedBy   int     `toml:"created_by"`
		Name        string  `toml:"name"`
		Lambda      float64 `toml:"lambda"`
		Price       float64 `toml:"price"`
		Description string  `toml:"description"`
	}

	type Insulation struct {
		Materials []Material `toml:"insulation"`
	}

	type Other struct {
		Materials []Material `toml:"other"`
	}

	type Wall struct {
		Materials []Material `toml:"wall"`
	}

	var insulations Insulation
	var others Other
	var walls Wall

	// Unmarshal the TOML data into the structs
	err = toml.Unmarshal(data, &insulations)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(data, &others)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(data, &walls)
	if err != nil {
		return err
	}

	// Extract the materials from the structs
	insulationMaterials := insulations.Materials
	otherMaterials := others.Materials
	wallMaterials := walls.Materials

	// Remove existing materials with the same name
	stmt := "DELETE FROM materials WHERE name IN ("
	names := make([]string, len(insulationMaterials)+len(otherMaterials)+len(wallMaterials))
	for i, material := range insulationMaterials {
		names[i] = fmt.Sprintf("'%s'", material.Name)
	}
	for i, material := range otherMaterials {
		names[i+len(insulationMaterials)] = fmt.Sprintf("'%s'", material.Name)
	}
	for i, material := range wallMaterials {
		names[i+len(insulationMaterials)+len(otherMaterials)] = fmt.Sprintf("'%s'", material.Name)
	}
	stmt += strings.Join(names, ", ") + ");"
	_, err = db.Exec(stmt)
	if err != nil {
		return err
	}

	// Create SQL statements for each material
	stmt = ""
	for _, material := range insulationMaterials {
		stmt += fmt.Sprintf("INSERT INTO materials (created_by, name, lambda, price, description) VALUES (%d, '%s', %f, %f, '%s');\n",
			material.CreatedBy, material.Name, material.Lambda, material.Price, material.Description)
	}
	for _, material := range otherMaterials {
		stmt += fmt.Sprintf("INSERT INTO materials (created_by, name, lambda, price, description) VALUES (%d, '%s', %f, %f, '%s');\n",
			material.CreatedBy, material.Name, material.Lambda, material.Price, material.Description)
	}
	for _, material := range wallMaterials {
		stmt += fmt.Sprintf("INSERT INTO materials (created_by, name, lambda, price, description) VALUES (%d, '%s', %f, %f, '%s');\n",
			material.CreatedBy, material.Name, material.Lambda, material.Price, material.Description)
	}

	// Execute the SQL statements
	_, err = db.Exec(stmt)
	if err != nil {
		return err
	}

	return nil
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

	stmt = `CREATE TABLE IF NOT EXISTS materials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_by INTEGER NOT NULL,
		name VARCHAR(64) NOT NULL,
		lambda REAL NOT NULL,
		price REAL NOT NULL,
		description VARCHAR(255) NULL,
		FOREIGN KEY(created_by) REFERENCES users(id)
	);`

	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}

	// Add items from materials.toml file
	err = addMaterialsFromTomlFile("./assets/data/materials.toml")
	if err != nil {
		log.Fatal(err)
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
