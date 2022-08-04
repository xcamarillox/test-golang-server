package main

import (
	"fmt"
	"os"
	"reto/awsAuxLib"
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
	PhotoURL         string  `json:"photoURL"`
	VerifiedURL      bool    `json:"verifiedURL"`
	PhotoExtension   string  `json:"photoExtension"`
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
		VerifiedURL:      false,
		PhotoURL:         "",
		PhotoExtension:   "",
	},
	{
		Id:               1,
		Make:             "ford",
		Model:            "bronco",
		Year:             2022,
		EngineCapacity:   2.5,
		Color:            "blanca",
		TransmissionType: "automatica",
		VerifiedURL:      false,
		PhotoURL:         "",
		PhotoExtension:   "",
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
			return 400, newCar
		}
	}

	if mode == "checkCarIndex" || mode == "checkRequestDataAndCarIndex" || mode == "checkCarIndexAndImage" {
		if carIndex < 0 {
			c.SendString("El ID enviado en la ruta es incorrecto o no fue encontrado.")
			return 404, newCar // Error code 404
		}
		if mode == "checkCarIndexAndImage" {
			if _, err := c.FormFile("photo"); err != nil {
				c.SendString("Debes enviar una archivo de imagen válido. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
				return 404, newCar
			}
		}
	}
	return 0, newCar //incoming data is ok and setted in car, 0 returned as no error code
}

func main() {
	app := fiber.New()

	awsAuxLib.S3.Region = "us-east-1"
	awsAuxLib.S3.NewSession(awsAuxLib.S3.Region)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		responseCode, newCar := getMeAReponseAndOrANewCar(c, "checkRequestData", 0)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		newCar.Id = getNewIntId()
		newCar.PhotoURL = ""
		newCar.PhotoExtension = ""
		newCar.VerifiedURL = false
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
	app.Put("/:id/client-upload", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars[carIndex].VerifiedURL = true
		c.SendString("La URL fue verificada con exito.")
		return c.SendStatus(202)

	})
	app.Put("/:id/set-photo/:mode", func(c *fiber.Ctx) error {
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
		fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + extension
		pathAndFile := fmt.Sprintf("/photos/%s", fileName)
		if c.Params("mode") == "client-upload" || c.Params("mode") == "server-upload" {
			var photoURL string
			c.SaveFile(file, "./public"+pathAndFile)
			if c.Params("mode") == "server-upload" {
				awsAuxLib.S3.UploadObject("./public"+pathAndFile, "levita-uploads-dev", fileName)
				photoURL, _ = awsAuxLib.S3.GetFileUrl("levita-uploads-dev", fileName)
				availableCars[carIndex].VerifiedURL = true
			} else {
				photoURL, _ = awsAuxLib.S3.GetAPresignedURL("./public"+pathAndFile, "levita-uploads-dev", fileName)
				availableCars[carIndex].VerifiedURL = false
			}
			os.Remove("./public" + pathAndFile)
			c.SendString("El recurso fue editado con exito. Se generó la siguiente URL:")
			c.SendString("URL del recurso:")
			c.SendString(photoURL)
			availableCars[carIndex].PhotoExtension = extension
			availableCars[carIndex].PhotoURL = photoURL
		} else {
			c.SendString("Method Not Allowed")
			return c.SendStatus(405)
		}
		return c.SendStatus(202)
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		carIndex, idInt := getIndexOfStringId(c.Params("id"))
		responseCode, carToEdit := getMeAReponseAndOrANewCar(c, "checkRequestDataAndCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		carToEdit.PhotoURL = availableCars[carIndex].PhotoURL
		carToEdit.VerifiedURL = availableCars[carIndex].VerifiedURL
		carToEdit.PhotoExtension = availableCars[carIndex].PhotoExtension
		availableCars[carIndex] = carToEdit
		availableCars[carIndex].Id = idInt
		c.SendString("El recurso fue editado con exito.")
		return c.SendStatus(202)

	})
	app.Delete("/:id/:remove-photo", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + availableCars[carIndex].PhotoExtension
		availableCars[carIndex].PhotoURL = ""
		availableCars[carIndex].VerifiedURL = false
		availableCars[carIndex].PhotoExtension = ""
		awsAuxLib.S3.DeleteObject(fileName, "levita-uploads-dev", fileName)
		c.SendString("El recurso fue eliminado con exito.")
		return c.SendStatus(200)
	})
	app.Delete("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, "checkCarIndex", carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		if availableCars[carIndex].PhotoURL != "" {
			fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + availableCars[carIndex].PhotoExtension
			awsAuxLib.S3.DeleteObject(fileName, "levita-uploads-dev", fileName)
		}
		availableCars = append(availableCars[:carIndex], availableCars[carIndex+1:]...)
		c.SendString("El recurso fue eliminado con exito.")
		return c.SendStatus(202)
	})

	//app.Static("/photos", "./public/photos")
	app.Listen(":3000")
}
