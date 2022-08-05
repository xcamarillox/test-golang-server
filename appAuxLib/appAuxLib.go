package appAuxLib

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	CheckCarIndex                      = 1
	CheckRequestData                   = 2
	CheckCarIndexAndImage              = 3
	CheckRequestDataAndCarIndex        = 4
	CheckCarIndexAndPhotoFileNameField = 5
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

func GetIndexOfIntId(id int, availableCars []CarSpecs) int {
	for index, car := range availableCars {
		if car.Id == id {
			return index
		}
	}
	return -1
}

func GetIndexOfStringId(id string, availableCars []CarSpecs) (int, int) {
	var carIndex int
	idInt, err := strconv.Atoi(id)
	if err == nil {
		carIndex = GetIndexOfIntId(idInt, availableCars)
		return carIndex, idInt
	}
	return -1, -1
}

func GetNewIntId(availableCars []CarSpecs) int {
	for index := range availableCars {
		if GetIndexOfIntId(index, availableCars) < 0 {
			return index
		}
	}
	return len(availableCars)
}

func GetPhotoFileExtension(filename string) (ext string, errorString string) {
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

func GetFileExtensionFromURL(url string) string {
	splitPath := strings.Split(url, "/")
	splitPathFileName := strings.Split(splitPath[len(splitPath)-1], ".")
	extension := splitPathFileName[len(splitPathFileName)-1]
	return extension
}

func GetMeAReponseAndOrANewCar(c *fiber.Ctx, mode int, carIndex int) (int, CarSpecs) {
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
			_, errorString = GetPhotoFileExtension(file.Filename)
		}
		if mode == CheckCarIndexAndPhotoFileNameField {
			field := c.FormValue("photoFileName")
			if field == "" {
				c.SendString("Debes anexar el campo photoFileName además del nombre y extensión de tu imágen. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
				return 400, newCar
			}
			_, errorString = GetPhotoFileExtension(field)
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
