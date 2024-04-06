package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kaloszer/insulationCalcHtmx/views/material_views"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/kaloszer/insulationCalcHtmx/models"
	"github.com/sujit-baniya/flash"
)

/********** Handlers for Material Views **********/

// Render List Page with success/error messages
func HandleMaterialViewList(c *fiber.Ctx) error {
	material := new(models.Material)
	material.CreatedBy = c.Locals("userId").(uint64)

	fm := fiber.Map{
		"type": "error",
	}

	materialsSlice, err := material.GetAllMaterials()
	if err != nil {
		fm["message"] = fmt.Sprintf("something went wrong: %s", err)

		return flash.WithError(c, fm).Redirect("/material/create")
	}

	tindex := material_views.MaterialIndex(materialsSlice)
	tlist := material_views.MaterialList(
		" | materials List",
		fromProtected,
		flash.Get(c),
		c.Locals("username").(string),
		tindex,
	)

	handler := adaptor.HTTPHandler(templ.Handler(tlist))

	return handler(c)
}

// Render Create Material Page with success/error messages
func HandleViewMaterialCreatePage(c *fiber.Ctx) error {

	if c.Method() == "POST" {
		fm := fiber.Map{
			"type": "error",
		}

		newMaterial := new(models.Material)
		newMaterial.CreatedBy = c.Locals("userId").(uint64)
		newMaterial.Name = strings.Trim(c.FormValue("title"), " ")
		newMaterial.Description = strings.Trim(c.FormValue("description"), " ")

		valueStr := c.FormValue("lambda") // This is a string.
		value, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		newMaterial.Lambda = value

		if _, err := newMaterial.CreateMaterial(); err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)

			return flash.WithError(c, fm).Redirect("/material/list")
		}

		return c.Redirect("/material/list")
	}

	cindex := material_views.CreateIndex()
	create := material_views.Create(
		" | Create a new material",
		fromProtected,
		flash.Get(c),
		c.Locals("username").(string),
		cindex,
	)

	handler := adaptor.HTTPHandler(templ.Handler(create))

	return handler(c)
}

// Render Edit Material Page with success/error messages
func HandleViewMaterialEditPage(c *fiber.Ctx) error {
	idParams, _ := strconv.Atoi(c.Params("id"))
	materialId := uint64(idParams)

	material := new(models.Material)
	material.ID = materialId
	material.CreatedBy = c.Locals("userId").(uint64)

	fm := fiber.Map{
		"type": "error",
	}

	recoveredMaterial, err := material.GetMaterialById()
	if err != nil {
		fm["message"] = fmt.Sprintf("something went wrong: %s", err)

		return flash.WithError(c, fm).Redirect("/material/list")
	}

	if c.Method() == "POST" {
		material.Name = strings.Trim(c.FormValue("title"), " ")
		material.Description = strings.Trim(c.FormValue("description"), " ")

		valueStr := c.FormValue("lambda") // This is a string.
		value, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		material.Lambda = value

		_, err = material.UpdateMaterial()
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)

			return flash.WithError(c, fm).Redirect("/material/list")
		}

		fm = fiber.Map{
			"type":    "success",
			"message": "Material successfully updated!!",
		}

		return flash.WithSuccess(c, fm).Redirect("/material/list")
	}

	uindex := material_views.UpdateIndex(recoveredMaterial)
	update := material_views.Update(
		fmt.Sprintf(" | Edit Material #%d", recoveredMaterial.ID),
		fromProtected,
		flash.Get(c),
		c.Locals("username").(string),
		uindex,
	)

	handler := adaptor.HTTPHandler(templ.Handler(update))

	return handler(c)
}

// Handler Remove Material
func HandleDeleteMaterial(c *fiber.Ctx) error {
	idParams, _ := strconv.Atoi(c.Params("id"))
	MaterialId := uint64(idParams)

	Material := new(models.Material)
	Material.ID = MaterialId
	Material.CreatedBy = c.Locals("userId").(uint64)

	fm := fiber.Map{
		"type": "error",
	}

	if err := Material.DeleteMaterial(); err != nil {
		fm["message"] = fmt.Sprintf("something went wrong: %s", err)

		return flash.WithError(c, fm).Redirect(
			"/material/list",
			fiber.StatusSeeOther,
		)
	}

	fm = fiber.Map{
		"type":    "success",
		"message": "Task successfully deleted!!",
	}

	return flash.WithSuccess(c, fm).Redirect("/material/list", fiber.StatusSeeOther)
}
