package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	IsBrandNew       bool    `json:"isBrandNew"`
	PhotoPath        string  `json:"photoPath"`
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
		IsBrandNew:       true,
		PhotoPath:        "",
	},
	{
		Id:               1,
		Make:             "ford",
		Model:            "bronco",
		Year:             2022,
		EngineCapacity:   2.5,
		Color:            "blanca",
		TransmissionType: "automatica",
		IsBrandNew:       true,
		PhotoPath:        "",
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
func getMeAReponseAndOrANewCar(c *fiber.Ctx, mode string, carIndex int) (int, CarSpecs) {
	newCar := CarSpecs{}
	if mode == "checkRequestData" || mode == "checkRequestDataAndCarIndex" {
		if err := c.BodyParser(&newCar); err != nil {
			c.SendString("La estructura de los datos recibidos es incorrecta.")
			return 400, newCar // Error code 400
		}
		if newCar.Year < 0 {
			c.SendString("El año de manufactura del automóvil no puede ser negativo.")
			return 400, newCar // Error code 400
		}
	}

	if mode == "checkCarIndex" || mode == "checkRequestDataAndCarIndex" || mode == "checkCarIndexAndImage" {
		if carIndex < 0 {
			c.SendString("El ID enviado en la ruta es incorrecto o no fue encontrado.")
			return 404, newCar // Error code 404
		}
		if mode == "checkCarIndexAndImage" {
			if _, err := c.FormFile("photo"); err != nil {
				c.SendString("Debes enviar una imagen válida.")
				return 404, newCar
			}
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
		responseCode, newCar := getMeAReponseAndOrANewCar(c, "checkRequestData", 0)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		newCar.Id = getNewIntId()
		newCar.PhotoPath = ""
		availableCars = append(availableCars, newCar)
		c.SendString("El recurso fue añadido con exito.")
		return c.SendStatus(201)
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
	})

	app.Put("/:id/set-photo", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndexAndImage", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		file, _ := c.FormFile("photo")
		splitStr := strings.Split(file.Filename, ".")
		extension := strings.ToLower(splitStr[len(splitStr)-1])
		if extension != "jpg" && extension != "jpeg" && extension != "png" && extension != "gif" || len(splitStr) < 2 {
			if len(splitStr) < 2 {
				c.SendString("Error con el nombre de archivo. El archivo debe tener nombre y extensión.")
			} else {
				c.SendString("Error en el tipo de archivo. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
			}
			return c.SendStatus(404)
		}
		fileName := availableCars[carIndex].Model + "_" + strconv.Itoa(availableCars[carIndex].Year) + "-" + strconv.Itoa(availableCars[carIndex].Id) + "." + extension
		pathAndFile := fmt.Sprintf("/photos/%s", fileName)
		availableCars[carIndex].PhotoPath = pathAndFile
		c.SaveFile(file, "./public"+pathAndFile)
		return c.SendStatus(202)
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		carIndex, idInt := getIndexOfStringId(c.Params("id"))
		responseCode, carToEdit := getMeAReponseAndOrANewCar(c, "checkRequestDataAndCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars[carIndex] = carToEdit
		availableCars[carIndex].Id = idInt
		c.SendString("El recurso fue editado con exito.")
		return c.SendStatus(202)

	})
	app.Delete("/:id/remove-photo", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		err := os.Remove("./public" + availableCars[carIndex].PhotoPath)
		if err != nil {
			c.SendString("Error al eliminar archivo.")
			return c.SendStatus(400)
		}
		availableCars[carIndex].PhotoPath = ""
		c.SendString("El recurso fue eliminado con exito.")
		return c.SendStatus(200)
	})
	app.Delete("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars = append(availableCars[:carIndex], availableCars[carIndex+1:]...)
		c.SendString("El recurso fue eliminado con exito.")
		return c.SendStatus(202)
	})

	app.Static("/photos", "./public/photos")
	app.Listen(":3000")
}
