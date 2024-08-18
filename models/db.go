package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

func GetMaterialsByIDs(ids string) ([]Material, error) {
	getConnection()

	// Split the ids string into a slice of strings
	idSlice := strings.Split(ids, ",")

	// Create a slice of interface{} to hold the id values
	args := make([]interface{}, len(idSlice))
	for i, id := range idSlice {
		args[i] = id
	}

	// Create a parameterized query with the correct number of placeholders
	query := fmt.Sprintf("SELECT id, created_by, name, description, lambda, price, thickness FROM materials WHERE id IN (%s)",
		strings.Join(strings.Split(strings.Repeat("?", len(idSlice)), ""), ","))

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying materials: %w", err)
	}
	defer rows.Close()

	// Slice to hold the results
	var materials []Material

	// Iterate over the rows and scan the results into Material structs
	for rows.Next() {
		var m Material
		err := rows.Scan(&m.ID, &m.CreatedBy, &m.Name, &m.Description, &m.Lambda, &m.Price, &m.Thickness)
		if err != nil {
			return nil, fmt.Errorf("error scanning material row: %w", err)
		}
		materials = append(materials, m)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over material rows: %w", err)
	}

	return materials, nil
}

func AddMaterial(material Material) error {

	stmt := `INSERT INTO materials (created_by, name, lambda, price, thickness, description)
		VALUES(?, ?, ?, ?, ?, ?);`

	log.Println(stmt, material.CreatedBy, material.Name, material.Lambda, material.Price, material.Thickness, material.Description)

	_, err := db.Exec(stmt, material.CreatedBy, material.Name, material.Lambda, material.Price, material.Thickness, material.Description)

	if err != nil {
		return fmt.Errorf("error adding material: %w", err)
	}

	return nil
}

func GetAllMaterials() ([]Material, error) {

	stmt := `SELECT id, created_by, name, description, lambda, price, thickness FROM materials;`
	log.Println(stmt)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("error querying materials: %w", err)
	}
	defer rows.Close()

	var materials []Material

	for rows.Next() {
		var m Material
		err := rows.Scan(&m.ID, &m.CreatedBy, &m.Name, &m.Description, &m.Lambda, &m.Price, &m.Thickness)
		if err != nil {
			return nil, fmt.Errorf("error scanning material row: %w", err)
		}
		materials = append(materials, m)
	}

	return materials, nil
}

/*
https://noties.io/blog/2019/08/19/sqlite-toggle-boolean/index.html
*/
