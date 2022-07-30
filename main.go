package main

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CarSpecs struct {
	Id               int
	Make             string
	Model            string
	Year             int
	EngineCapacity   float32 //enlitros
	Color            string
	TransmissionType string // manual / automatica
	FuelType         string //gasolina / diesel
	IsBrandNew       bool
}

var availableCars = []CarSpecs{
	CarSpecs{
		Id:               0,
		Make:             "chevrolet",
		Model:            "trax",
		Year:             2015,
		EngineCapacity:   1.8,
		Color:            "roja",
		TransmissionType: "automatica",
		FuelType:         "gasolina",
		IsBrandNew:       true,
	},
	CarSpecs{
		Id:               1,
		Make:             "ford",
		Model:            "bronco",
		Year:             2022,
		EngineCapacity:   2.5,
		Color:            "blanca",
		TransmissionType: "automatica",
		FuelType:         "gasolina",
		IsBrandNew:       true,
	},
}

func checkIfCarIsFound(id int) (bool, int) {
	for index, car := range availableCars {
		if car.Id == id {
			return true, index
		}
	}
	return false, 0
}

func checkIfIdIsCorrect(id string) (int, error) {
	idInt, err := strconv.Atoi(id)
	return idInt, err
}

func main() {

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		var existingId bool
		var carIndex int
		idInt, err := checkIfIdIsCorrect(c.Params("id"))
		if err == nil {
			existingId, carIndex = checkIfCarIsFound(idInt)
		}
		if err == nil && existingId {
			return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
		}
		return c.SendStatus(404)
	})
	app.Put("/:id", func(c *fiber.Ctx) error {
		var existingId bool
		var carIndex int
		idInt, err := checkIfIdIsCorrect(c.Params("id"))
		if err == nil {
			existingId, carIndex = checkIfCarIsFound(idInt)
		}
		if err == nil && existingId {
			//return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
			fmt.Println(carIndex)
		}
		return c.SendStatus(404)
	})
	app.Delete("/:id", func(c *fiber.Ctx) error {
		var existingId bool
		var carIndex int
		idInt, err := checkIfIdIsCorrect(c.Params("id"))
		if err == nil {
			existingId, carIndex = checkIfCarIsFound(idInt)
		}
		if err == nil && existingId {
			//return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
			fmt.Println(carIndex)
		}
		return c.SendStatus(404)
	})
	app.Listen(":3000")
}
