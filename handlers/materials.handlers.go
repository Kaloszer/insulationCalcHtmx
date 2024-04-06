package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/kaloszer/insulationcalchtmx/models"
	"github.com/sujit-baniya/flash"

	"views/material_views"
)

/********** Handlers for Material Views **********/

// Render List Page with success/error messages
func HandleMaterialViewList(c *fiber.Ctx) error {
	Material := new(models.Material)
	Material.CreatedBy = c.Locals("userId").(uint64)

	fm := fiber.Map{
		"type": "error",
	}

	MaterialsSlice, err := Material.GetAllMaterials()
	if err != nil {
		fm["message"] = fmt.Sprintf("something went wrong: %s", err)

		return flash.WithError(c, fm).Redirect("/Material/list")
	}

	tindex := material_views.MaterialIndex(MaterialsSlice)
	tlist := material_views.MaterialList(
		" | Materials List",
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
		newMaterial.Lambda = c.QueryFloat("lambda", 0.20)

		if _, err := newMaterial.CreateMaterial(); err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)

			return flash.WithError(c, fm).Redirect("/Material/list")
		}

		return c.Redirect("/Material/list")
	}

	cindex := Material_views.CreateIndex()
	create := Material_views.Create(
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
	MaterialId := uint64(idParams)

	Material := new(models.Material)
	Material.ID = MaterialId
	Material.CreatedBy = c.Locals("userId").(uint64)

	fm := fiber.Map{
		"type": "error",
	}

	recoveredMaterial, err := Material.GetMaterialById()
	if err != nil {
		fm["message"] = fmt.Sprintf("something went wrong: %s", err)

		return flash.WithError(c, fm).Redirect("/Material/list")
	}

	if c.Method() == "POST" {
		Material.Name = strings.Trim(c.FormValue("title"), " ")
		Material.Description = strings.Trim(c.FormValue("description"), " ")
		Material.Lambda = c.QueryFloat("lambda")

		_, err := Material.UpdateMaterial()
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)

			return flash.WithError(c, fm).Redirect("/Material/list")
		}

		fm = fiber.Map{
			"type":    "success",
			"message": "Material successfully updated!!",
		}

		return flash.WithSuccess(c, fm).Redirect("/Material/list")
	}

	uindex := Material_views.UpdateIndex(recoveredMaterial)
	update := Material_views.Update(
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
			"/Material/list",
			fiber.StatusSeeOther,
		)
	}

	fm = fiber.Map{
		"type":    "success",
		"message": "Task successfully deleted!!",
	}

	return flash.WithSuccess(c, fm).Redirect("/Material/list", fiber.StatusSeeOther)
}
