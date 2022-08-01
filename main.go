package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CarSpecs struct {
	Id               int     `json:"id"`
	Make             string  `json:"make"`
	Model            string  `json:"model"`
	Year             int     `json:"year"`
	EngineCapacity   float32 `json:"engineCapacity"` //enlitros
	Color            string  `json:"color"`
	TransmissionType string  `json:"transmissionType"` // manual / automatica
	FuelType         string  `json:"fuelType"`         //gasolina / diesel
	IsBrandNew       bool    `json:"isBrandNew"`
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
func getMeAReponseAndANewCarWithCarData(c *fiber.Ctx, mode string, carIndex int) (int, CarSpecs) {
	newCar := CarSpecs{}
	if mode == "checkDataGetMeACar" || mode == "checkDataAndCarIndexGetMeACar" {
		if err := c.BodyParser(&newCar); err != nil {
			c.SendString("La estructura de los datos recibidos es incorrecta.")
			return 400, newCar // Error code 400
		}
		if newCar.Year < 0 {
			c.SendString("El año de manufactura del automóvil no puede ser negativo.")
			return 400, newCar // Error code 400
		}
	}
	if mode == "checkCarIndex" || mode == "checkDataAndCarIndexGetMeACar" {
		if carIndex < 0 {
			c.SendString("El ID enviado en la ruta es incorrecto o no fue encontrado.")
			return 404, newCar // Error code 404
		}
	}
	return 0, newCar //incoming data is ok and setted in car, 0 returned as no error code
}
func main() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		responseCode, newCar := getMeAReponseAndANewCarWithCarData(c, "checkDataGetMeACar", 0)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		newCar.Id = getNewIntId()
		availableCars = append(availableCars, newCar)
		c.SendString("El recurso fue añadido con exito.")
		return c.SendStatus(201)
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndANewCarWithCarData(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
	})
	app.Put("/:id", func(c *fiber.Ctx) error {
		carIndex, idInt := getIndexOfStringId(c.Params("id"))
		responseCode, carToEdit := getMeAReponseAndANewCarWithCarData(c, "checkDataAndCarIndexGetMeACar", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars[carIndex] = carToEdit
		availableCars[carIndex].Id = idInt
		c.SendString("El recurso fue editado con exito.")
		return c.SendStatus(202)

	})
	app.Delete("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndANewCarWithCarData(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars = append(availableCars[:carIndex], availableCars[carIndex+1:]...)
		c.SendString("El recurso fue eliminado con exito.")
		return c.SendStatus(202)
	})
	app.Listen(":3000")
}
