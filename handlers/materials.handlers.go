package handlers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/kaloszer/insulationCalcHtmx/models"
	"github.com/kaloszer/insulationCalcHtmx/views/material_views"
	"github.com/olekukonko/tablewriter"
	"github.com/sujit-baniya/flash"
)

// HandleViewMaterialCreatePage handler
func HandleViewMaterialCreatePage(c *fiber.Ctx) error {
	if c.Method() == "POST" {
		return c.Redirect("/material/list")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Lambda", "Price", "Description"})

	materials, err := models.LoadMaterialsFromTOML("materials.toml")
	if err != nil {
		log.Printf("Error reading materials from TOML file: %v", err)
		return flash.WithError(c, fiber.Map{
			"type":    "error",
			"message": "Error reading materials from TOML file",
		}).Redirect("/material/list")
	}

	for _, material := range materials {
		table.Append([]string{
			material.Name,
			fmt.Sprintf("%f", material.Lambda),
			fmt.Sprintf("%f", material.Price),
			material.Description,
		})
	}

	table.SetCaption(true, "Materials")
	table.Render()

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
		value, err := strconv.ParseFloat(valueStr, 32)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		material.Lambda = float32(value)

		valuePriceStr := c.FormValue("price") // This is a string.
		value, err = strconv.ParseFloat(valuePriceStr, 32)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		material.Price = float32(value)

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

// Search Material
func HandleViewMaterialSearch(c *fiber.Ctx) error {
	search := new(models.Search)

	fm := fiber.Map{
		"type": "error",
	}

	if c.Method() == "POST" {
		search.Name = strings.Trim(c.FormValue("name"), " ")

		valueStr := c.FormValue("lambda") // This is a string.
		value, err := strconv.ParseFloat(valueStr, 32)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		search.Lambda = float32(value)

		valuePriceStr := c.FormValue("price") // This is a string.
		value, err = strconv.ParseFloat(valuePriceStr, 32)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		search.Price = float32(value)

		material := new(models.Material)

		materialsSlice, err := material.SearchMaterial(*search)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)

			return flash.WithError(c, fm).Redirect("/material/list")
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

	return c.Redirect("/material/list")

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
