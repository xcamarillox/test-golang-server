package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type CarSpecs struct {
	make             string
	model            string
	year             int
	engineCapacity   float32 //enlitros
	color            string
	transmissionType string // manual / automatica
	fuelType         string //gasolina / diesel
	isBrandNew       bool
}

func main() {
	availableCars := []CarSpecs{
		CarSpecs{
			make:             "chevrolet",
			model:            "trax",
			year:             2015,
			engineCapacity:   1.8,
			color:            "roja",
			transmissionType: "automatica",
			fuelType:         "gasolina",
			isBrandNew:       true,
		},
		CarSpecs{
			make:             "ford",
			model:            "bronco",
			year:             2022,
			engineCapacity:   2.5,
			color:            "blanca",
			transmissionType: "automatica",
			fuelType:         "gasolina",
			isBrandNew:       true,
		},
	}
	fmt.Println(availableCars)

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":3000")
}
