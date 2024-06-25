package handlers

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
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
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		material.Lambda = float64(value)

		valuePriceStr := c.FormValue("price") // This is a string.
		value, err = strconv.ParseFloat(valuePriceStr, 64)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		material.Price = float64(value)

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
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		search.Lambda = float64(value)

		valuePriceStr := c.FormValue("price") // This is a string.
		value, err = strconv.ParseFloat(valuePriceStr, 64)
		if err != nil {
			fm["message"] = fmt.Sprintf("something went wrong: %s", err)
			return flash.WithError(c, fm).Redirect("/material/list")
		}
		search.Price = float64(value)

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

// HandleInsulationCalculatorPage renders the insulation calculator page
// func HandleInsulationCalculatorPage(c *fiber.Ctx) error {
// 	materials, err := models.LoadMaterialsFromTOML("./assets/data/materials.toml")
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).SendString("Error loading materials: " + err.Error())
// 	}
// 	// no clue how to finish this // TODO
// 	return adaptor.HTTPHandlerFunc(material_views.InsulationCalculatorPage(materials))()
// }

func HandleInsulationCalculatorPage(c *fiber.Ctx) error {
	materials, err := models.LoadMaterialsFromTOML("./assets/data/materials.toml")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error loading materials: " + err.Error())
	}

	handler := adaptor.HTTPHandler(templ.Handler(material_views.InsulationCalculatorPage(materials)))

	return handler(c)
}

// HandleCalculateInsulation handles the insulation calculation request
func HandleCalculateInsulation(c *fiber.Ctx) error {
	// Parse input parameters
	wallType := c.FormValue("wall-type")
	desiredUValue, err := strconv.ParseFloat(c.FormValue("desired-u-value"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid desired U-value")
	}

	// Get selected materials
	materialIDs := c.FormValue("insulation-materials")
	materials, err := models.GetMaterialsByIDs(materialIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching materials: " + err.Error())
	}

	// Perform optimization
	result := optimizeInsulation(wallType, desiredUValue, materials)

	// Render the result using the templ component
	return material_views.InsulationResult(result).Render(c.Context(), c.Response().BodyWriter())
}

func optimizeInsulation(wallType string, desiredUValue float64, materials []models.Material) models.InsulationResult {
	result := models.InsulationResult{
		Layers:      []models.InsulationLayer{},
		TotalUValue: math.Inf(1),
		TotalCost:   0,
	}

	// Sort materials by their insulation efficiency (lambda)
	sort.Slice(materials, func(i, j int) bool {
		return materials[i].Lambda < materials[j].Lambda
	})

	currentUValue := 0.0
	for currentUValue < desiredUValue {
		bestMaterial := materials[0]
		bestThickness := 0.0
		bestCost := math.Inf(1)

		for _, material := range materials {
			thickness := calculateRequiredThickness(material.Lambda, desiredUValue-currentUValue)
			cost := thickness * material.Price / 1000 // Assuming price is per m³ and thickness is in mm

			if cost < bestCost {
				bestMaterial = material
				bestThickness = thickness
				bestCost = cost
			}
		}

		layer := models.InsulationLayer{
			Material:  bestMaterial,
			Thickness: bestThickness,
			UValue:    1 / (bestThickness / 1000 / bestMaterial.Lambda),
		}
		result.Layers = append(result.Layers, layer)
		result.TotalCost += bestCost
		currentUValue += layer.UValue
	}

	result.TotalUValue = currentUValue
	return result
}

func calculateRequiredThickness(lambda, targetUValue float64) float64 {
	return 1 / (targetUValue * lambda) * 1000 // Convert to mm
}
