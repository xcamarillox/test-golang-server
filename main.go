package main

import (
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

func main() {

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})

	app.Listen(":3000")
}
