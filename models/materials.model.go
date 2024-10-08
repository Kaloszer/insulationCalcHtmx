package models

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Material struct {
	ID          uint64  `json:"id" toml:"id"`
	CreatedBy   uint64  `json:"created_by" toml:"created_by"`
	Name        string  `json:"name" toml:"name"`
	Description string  `json:"description,omitempty" toml:"description"`
	Lambda      float64 `json:"lambda" toml:"lambda"`
	Price       float64 `json:"price,omitempty" toml:"price"`
	Thickness   float64 `json:"thickness" toml:"thickness"`
	Type        string  `json:"type" toml:"type"`
}

// New structs for insulation calculation
type InsulationLayer struct {
	Material  Material `json:"material"`
	Thickness float64  `json:"thickness"`
	UValue    float64  `json:"u_value"`
}

type InsulationResult struct {
	Layers      []InsulationLayer `json:"layers"`
	TotalUValue float64           `json:"total_u_value"`
	TotalCost   float64           `json:"total_cost"`
}

// TOMLData represents the structure of your TOML file
type TOMLData struct {
	Insulation []Material `toml:"insulation"`
	Other      []Material `toml:"other"`
	Wall       []Material `toml:"wall"`
}

// LoadMaterialsFromTOML loads materials from a TOML file
func LoadMaterialsFromTOML(filename string) ([]Material, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data TOMLData
	if err := toml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	// Combine all materials into a single slice
	var materials []Material
	materials = append(materials, data.Insulation...)
	materials = append(materials, data.Other...)
	materials = append(materials, data.Wall...)

	return materials, nil
}

type Search struct {
	Name   string  `json:"name"`
	Lambda float64 `json:"lambda"`
	Price  float64 `json:"price"`
}

func (t *Material) GetAllMaterials() ([]Material, error) {
	query := fmt.Sprintf("SELECT id, name, description, lambda FROM materials WHERE created_by IN (%d, 1337) ORDER BY name DESC", t.CreatedBy)

	rows, err := db.Query(query)
	if err != nil {
		return []Material{}, err
	}
	// We close the resource
	defer rows.Close()

	Materials := []Material{}
	for rows.Next() {
		rows.Scan(&t.ID, &t.Name, &t.Description, &t.Lambda)

		Materials = append(Materials, *t)
	}

	return Materials, nil
}

func (t *Material) GetMaterialById() (Material, error) {

	query := `SELECT id, name, description, lambda FROM materials
		WHERE created_by = ? AND id=?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return Material{}, err
	}

	defer stmt.Close()

	var recoveredMaterial Material
	err = stmt.QueryRow(
		t.CreatedBy, t.ID,
	).Scan(
		&recoveredMaterial.ID,
		&recoveredMaterial.Name,
		&recoveredMaterial.Description,
		&recoveredMaterial.Lambda,
	)
	if err != nil {
		return Material{}, err
	}

	return recoveredMaterial, nil
}

func (t *Material) CreateMaterial() (Material, error) {

	query := `INSERT INTO materials (created_by, name, description, lambda, thickness, price)
    	VALUES(?, ?, ?, ?, ?) RETURNING id, created_by, name, description, thickness, lambda, price`

	stmt, err := db.Prepare(query)
	if err != nil {
		return Material{}, err
	}

	defer stmt.Close()

	var newMaterial Material
	err = stmt.QueryRow(
		t.CreatedBy,
		t.Name,
		t.Description,
		t.Lambda,
		t.Price,
	).Scan(
		&newMaterial.ID,
		&newMaterial.CreatedBy,
		&newMaterial.Name,
		&newMaterial.Description,
		&newMaterial.Lambda,
		&newMaterial.Price,
	)
	if err != nil {
		return Material{}, err
	}

	/* if i, err := result.RowsAffected(); err != nil || i != 1 {
		return errors.New("error: an affected row was expected")
	} */

	return newMaterial, nil
}

func (t *Material) UpdateMaterial() (Material, error) {

	if t.CreatedBy == 1 {
		return Material{}, errors.New("you cant update a system defined material 😭")
	}

	query := `UPDATE materials SET name = ?,  description = ?, status = ?, lambda = ?
		WHERE created_by = ? AND id=? RETURNING id, name, description, lambda`

	stmt, err := db.Prepare(query)
	if err != nil {
		return Material{}, err
	}

	defer stmt.Close()

	var updatedMaterial Material
	err = stmt.QueryRow(
		t.Name,
		t.Description,
		t.Lambda,
		t.CreatedBy,
		t.ID,
	).Scan(
		&updatedMaterial.ID,
		&updatedMaterial.Name,
		&updatedMaterial.Description,
		&updatedMaterial.Lambda,
	)
	if err != nil {
		return Material{}, err
	}

	return updatedMaterial, nil
}

func (t *Material) DeleteMaterial() error {

	if t.CreatedBy == 1 {
		return errors.New("you cant delete a system defined material 😭")
	}

	query := `DELETE FROM materials
		WHERE created_by = ? AND id=?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(t.CreatedBy, t.ID)
	if err != nil {
		return err
	}

	if i, err := result.RowsAffected(); err != nil || i != 1 {
		return errors.New("an affected row was expected")
	}

	return nil
}

func (t *Material) SearchMaterial(search Search) ([]Material, error) {
	query := "SELECT id, name, description, lambda FROM materials WHERE created_by IN (?, 1337)"

	args := []interface{}{t.CreatedBy}

	if search.Name != "" {
		query += " AND name ILIKE ?"
		args = append(args, "%"+search.Name+"%")
	}

	if search.Lambda != 0 {
		query += " AND lambda < ?"
		args = append(args, search.Lambda)
	}

	if search.Price != 0 {
		query += " AND price < ?"
		args = append(args, search.Price)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	materials := []Material{}
	for rows.Next() {
		var material Material
		err := rows.Scan(&material.ID, &material.Name, &material.Description, &material.Lambda)
		if err != nil {
			return nil, err
		}
		materials = append(materials, material)
	}

	return materials, nil
}
