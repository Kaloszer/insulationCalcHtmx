package models

import (
	"errors"
	"fmt"
)

type Material struct {
	ID          uint64 `json:"id"`
	CreatedBy   uint64 `json:"created_by"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Lambda      uint64 `json:"lambda"`
}

func (t *Material) GetAllMaterials() ([]Material, error) {
	query := fmt.Sprintf("SELECT id, name, description, lambda FROM materials WHERE created_by = %d ORDER BY name DESC", t.CreatedBy)

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

	query := `INSERT INTO Materials (created_by, name, description, lambda)
		VALUES(?, ?, ?, ?) RETURNING *`

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
	).Scan(
		&newMaterial.ID,
		&newMaterial.CreatedBy,
		&newMaterial.Name,
		&newMaterial.Description,
		&newMaterial.Lambda,
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

	query := `UPDATE Materials SET name = ?,  description = ?, status = ?, lambda = ?
		WHERE created_by = ? AND id=? RETURNING id, name, description, status`

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
