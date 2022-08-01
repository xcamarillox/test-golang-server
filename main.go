package main

import (
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
	{
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
	{
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

func getIndexOfIntId(id int) int {
	for index, car := range availableCars {
		if car.Id == id {
			return index
		}
	}
	return -1
}

func getIndexOfStringId(id string) (int, int) {
	var carIndex int
	idInt, err := strconv.Atoi(id)
	if err == nil {
		carIndex = getIndexOfIntId(idInt)
		return carIndex, idInt
	}
	return -1, -1
}

func getNewIntId() int {
	for index := range availableCars {
		if getIndexOfIntId(index) < 0 {
			return index
		}
	}
	return len(availableCars)
}

func main() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		newCar := CarSpecs{}
		if err := c.BodyParser(&newCar); err != nil {
			return c.SendStatus(400)
		}
		newCar.Id = getNewIntId()
		availableCars = append(availableCars, newCar)
		return c.SendStatus(202)
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		if carIndex < 0 {
			return c.SendStatus(404)
		}
		return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
	})
	app.Put("/:id", func(c *fiber.Ctx) error {
		carIndex, idInt := getIndexOfStringId(c.Params("id"))
		if carIndex < 0 {
			return c.SendStatus(404)
		}
		carToEdit := CarSpecs{}
		if err := c.BodyParser(&carToEdit); err != nil {
			return c.SendStatus(400)
		}
		availableCars[carIndex] = carToEdit
		availableCars[carIndex].Id = idInt
		return c.SendStatus(202)

	})
	app.Delete("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		if carIndex < 0 {
			return c.SendStatus(404)
		}
		availableCars = append(availableCars[:carIndex], availableCars[carIndex+1:]...)
		return c.SendStatus(202)
	})
	app.Listen(":3000")
}
