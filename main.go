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
}

const (
	CheckCarIndex                      = 1
	CheckRequestData                   = 2
	CheckCarIndexAndImage              = 3
	CheckRequestDataAndCarIndex        = 4
	CheckCarIndexAndPhotoFileNameField = 5
)

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

func getPhotoFileExtension(filename string) (ext string, errorString string) {
	splitFileName := strings.Split(filename, ".")
	extension := strings.ToLower(splitFileName[len(splitFileName)-1])
	if extension != "jpg" && extension != "jpeg" && extension != "png" && extension != "gif" || len(splitFileName) < 2 {
		if len(splitFileName) < 2 {
			return "", "Error con el nombre de archivo. El archivo debe tener nombre y extensión."
		} else {
			return "", "Error en el tipo de archivo. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente."
		}
	}
	return extension, ""
}

func getFileExtensionFromURL(url string) string {
	splitPath := strings.Split(url, "/")
	splitPathFileName := strings.Split(splitPath[len(splitPath)-1], ".")
	extension := splitPathFileName[len(splitPathFileName)-1]
	return extension
}

func getMeAReponseAndOrANewCar(c *fiber.Ctx, mode int, carIndex int) (int, CarSpecs) {
	newCar := CarSpecs{}
	if mode == CheckRequestData || mode == CheckRequestDataAndCarIndex {
		if err := c.BodyParser(&newCar); err != nil {
			c.SendString("La estructura de los datos recibidos es incorrecta.")
			return 400, newCar // Error code 400
		}
		if newCar.Year < 0 {
			c.SendString("El año de manufactura del automóvil no puede ser negativo.")
			return 400, newCar
		}
	}
	if mode == CheckCarIndex || mode == CheckRequestDataAndCarIndex || mode == CheckCarIndexAndImage || mode == CheckCarIndexAndPhotoFileNameField {
		var errorString string
		if carIndex < 0 {
			c.SendString("El ID enviado en la ruta es incorrecto o no fue encontrado.")
			return 404, newCar // Error code 404
		}
		if mode == CheckCarIndexAndImage {
			file, err := c.FormFile("photoFile")
			if err != nil {
				c.SendString("Debes anexar un archivo de imagen válido. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
				return 404, newCar
			}
			_, errorString = getPhotoFileExtension(file.Filename)
		}
		if mode == CheckCarIndexAndPhotoFileNameField {
			field := c.FormValue("photoFileName")
			if field == "" {
				c.SendString("Debes anexar el campo photoFileName además del nombre y extensión de tu imágen. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
				return 400, newCar
			}
			_, errorString = getPhotoFileExtension(field)
		}
		if mode == CheckCarIndexAndImage || mode == CheckCarIndexAndPhotoFileNameField {
			if errorString != "" {
				c.SendString(errorString)
				return 400, newCar
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
		responseCode, newCar := getMeAReponseAndOrANewCar(c, CheckRequestData, 0)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		newCar.Id = getNewIntId()
		newCar.PhotoURL = ""
		newCar.VerifiedURL = false
		availableCars = append(availableCars, newCar)
		c.SendString("El recurso fue añadido con exito.")
		return c.SendStatus(201)
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, CheckCarIndex, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
	})
	app.Put("/:id/client-upload", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, CheckCarIndex, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars[carIndex].VerifiedURL = true
		c.SendString("La URL fue verificada con exito.")
		return c.SendStatus(202)

	})
	app.Put("/:id/set-photo/:mode", func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		if c.Params("mode") == "client-upload" || c.Params("mode") == "server-upload" {
			var photoURL string
			var photoFileExtension string
			if c.Params("mode") == "server-upload" {
				responseCode, _ := getMeAReponseAndOrANewCar(c, CheckCarIndexAndImage, carIndex)
				if responseCode != 0 {
					return c.SendStatus(responseCode)
				}
				photoFile, _ := c.FormFile("photoFile")
				photoFileExtension, _ = getPhotoFileExtension(photoFile.Filename)
				fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + photoFileExtension
				pathAndFile := fmt.Sprintf("/photos/%s", fileName)
				c.SaveFile(photoFile, "./public"+pathAndFile)
				awsAuxLib.S3.UploadObject("./public"+pathAndFile, "levita-uploads-dev", fileName)
				photoURL, _ = awsAuxLib.S3.GetFileUrl("levita-uploads-dev", fileName)
				availableCars[carIndex].VerifiedURL = true
				os.Remove("./public" + pathAndFile)
			} else {
				responseCode, _ := getMeAReponseAndOrANewCar(c, CheckCarIndexAndPhotoFileNameField, carIndex)
				if responseCode != 0 {
					return c.SendStatus(responseCode)
				}
				photoFileNameFormValue := c.FormValue("photoFileName")
				photoFileExtension, _ = getPhotoFileExtension(photoFileNameFormValue)
				fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + photoFileExtension
				photoURL, _ = awsAuxLib.S3.GetAPresignedURL("levita-uploads-dev", fileName)
				availableCars[carIndex].VerifiedURL = false
			}
			c.SendString("El recurso fue editado con exito. Se generó la siguiente URL:\n" + photoURL)
			availableCars[carIndex].PhotoURL = strings.Split(photoURL, "?")[0]
		} else {
			c.SendString("Method Not Allowed")
			return c.SendStatus(405)
		}
		return c.SendStatus(202)
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		carIndex, idInt := getIndexOfStringId(c.Params("id"))
		responseCode, carToEdit := getMeAReponseAndOrANewCar(c, CheckRequestDataAndCarIndex, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		carToEdit.PhotoURL = availableCars[carIndex].PhotoURL
		carToEdit.VerifiedURL = availableCars[carIndex].VerifiedURL
		availableCars[carIndex] = carToEdit
		availableCars[carIndex].Id = idInt
		c.SendString("El recurso fue editado con exito.")
		return c.SendStatus(202)

	})

	var deleteHandler = func(c *fiber.Ctx) error {
		carIndex, _ := getIndexOfStringId(c.Params("id"))
		responseCode, _ := getMeAReponseAndOrANewCar(c, CheckCarIndex, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		if availableCars[carIndex].PhotoURL != "" {
			fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + getFileExtensionFromURL(availableCars[carIndex].PhotoURL)
			awsAuxLib.S3.DeleteObject("levita-uploads-dev", fileName)
		}
		availableCars[carIndex].PhotoURL = ""
		availableCars[carIndex].VerifiedURL = false
		route := c.Route()
		if route.Path == "/:id" {
			availableCars = append(availableCars[:carIndex], availableCars[carIndex+1:]...)
		}
		c.SendString("El recurso fue eliminado con exito.")
		return c.SendStatus(202)
	}

	app.Delete("/:id/remove-photo", deleteHandler)
	app.Delete("/:id", deleteHandler)

	//app.Static("/photos", "./public/photos")
	app.Listen(":3000")
}
