package models

import (
	"errors"
	"fmt"
)

type Material struct {
	ID          uint64  `json:"id"`
	CreatedBy   uint64  `json:"created_by"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Lambda      float32 `json:"lambda"`
	Price       float32 `json:"price,omitempty"`
}

type Search struct {
	Name   string  `json:"name"`
	Lambda float32 `json:"lambda"`
	Price  float32 `json:"price"`
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

	query := `INSERT INTO materials (created_by, name, description, lambda, price)
    	VALUES(?, ?, ?, ?, ?) RETURNING id, created_by, name, description, lambda, price`

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
