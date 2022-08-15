package reqHandlers

import (
	"fmt"
	"os"
	"reto/appAuxLib"
	"reto/awsAuxLib"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var availableCars = []appAuxLib.CarSpecs{}

var SetTestData = func(testData []appAuxLib.CarSpecs) {
	availableCars = testData
}

var GetRootHandler = func(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(availableCars)
}

var PostRootHandler = func(c *fiber.Ctx) error {
	responseCode, newCar := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckRequestData, 0)
	if responseCode != 0 {
		return c.SendStatus(responseCode)
	}
	newCar.Id = appAuxLib.GetNewIntId(availableCars)
	newCar.PhotoURL = ""
	newCar.VerifiedURL = false
	availableCars = append(availableCars, newCar)
	availableCars = appAuxLib.GetCarSpecsWithAPIData(availableCars) //Cylinders. FuelType
	c.SendString("El recurso fue añadido con exito.")
	return c.SendStatus(201)
}

var GetExportHandler = func(c *fiber.Ctx) error {
	f := appAuxLib.GetANewExcelizeFileOfCarSpecsSlice(availableCars)
	if err := f.SaveAs("./public/temp/BookTemp1.xlsx"); err != nil {
		fmt.Println(err)
	}
	c.SendFile("./public/temp/BookTemp1.xlsx")
	os.Remove("./public/temp/BookTemp1.xlsx")
	return c.SendStatus(202)
}

var PostImportHandler = func(c *fiber.Ctx) error {
	responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckExcelFile, 0)
	if responseCode != 0 {
		return c.SendStatus(responseCode)
	}
	file, _ := c.FormFile("excelFile")
	fileAndPath := "./public/temp/" + file.Filename
	c.SaveFile(file, fileAndPath)
	defer os.Remove("./public/temp/" + file.Filename)
	var err error
	var cellsWithErr []string
	availableCars, cellsWithErr, err = appAuxLib.ImportDataFromExcelFile(fileAndPath, availableCars)
	availableCars = appAuxLib.GetCarSpecsWithAPIData(availableCars)
	if err != nil {
		c.SendString("Error al importar los datos.")
		return c.SendStatus(400)
	}
	if len(cellsWithErr) != 0 {
		var leyenda string
		if len(cellsWithErr) == 1 {
			leyenda = cellsWithErr[0]
		}
		if len(cellsWithErr) > 1 {
			for i := range cellsWithErr {
				leyenda = leyenda + cellsWithErr[i]
				if i < len(cellsWithErr)-2 {
					leyenda = leyenda + ", "
					continue
				}
				if i == len(cellsWithErr)-2 {
					leyenda = leyenda + " y "
				}
			}
		}
		UrlOfFile, _ := appAuxLib.GetURLFileWithMarkedErrors(fileAndPath, cellsWithErr)
		UrlOfFile = strings.Split(UrlOfFile, "?")[0]
		c.SendString("Se importaron algunos datos, aunque se tuvieron problemas con las siguientes celdas:\n" + leyenda + "\nThe URL file:\n" + UrlOfFile)
		return c.SendStatus(202)
	}
	c.SendString("Los datos en el fichero de Excel han sido importados exitosamente.")
	return c.SendStatus(202)
}

var GetIdHandler = func(c *fiber.Ctx) error {
	carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
	responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndex, carIndex)
	if responseCode != 0 {
		return c.SendStatus(responseCode)
	}
	return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
}

var PutIdClientUploadHandler = func(c *fiber.Ctx) error {
	carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
	responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndex, carIndex)
	if responseCode != 0 {
		return c.SendStatus(responseCode)
	}
	availableCars[carIndex].VerifiedURL = true
	c.SendString("La URL fue verificada con exito.")
	return c.SendStatus(202)

}

var PutIdSetPhotoModeHandler = func(c *fiber.Ctx) error {
	carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
	if c.Params("mode") != "client" && c.Params("mode") != "server" {
		c.SendString("Method Not Allowed")
		return c.SendStatus(405)
	}
	var photoURL string
	var photoFileExtension string
	if c.Params("mode") == "server" {
		responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndexAndImage, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		photoFile, _ := c.FormFile("photoFile")
		photoFileExtension, _ = appAuxLib.GetPhotoFileExtension(photoFile.Filename)
		fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + photoFileExtension
		pathAndFile := fmt.Sprintf("/photos/%s", fileName)
		c.SaveFile(photoFile, "./public"+pathAndFile)
		awsAuxLib.S3.UploadObject("./public"+pathAndFile, "levita-uploads-dev", fileName)
		photoURL, _ = awsAuxLib.S3.GetTemporalUrl("levita-uploads-dev", fileName)
		availableCars[carIndex].VerifiedURL = true
		os.Remove("./public" + pathAndFile)
	}
	if c.Params("mode") == "client" {
		responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndexAndPhotoFileNameField, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		photoFileNameFormValue := c.FormValue("photoFileName")
		photoFileExtension, _ = appAuxLib.GetPhotoFileExtension(photoFileNameFormValue)
		fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + photoFileExtension
		photoURL, _ = awsAuxLib.S3.GetAPresignedURL("levita-uploads-dev", fileName)
		availableCars[carIndex].VerifiedURL = false
	}
	availableCars[carIndex].PhotoURL = strings.Split(photoURL, "?")[0]
	c.SendString("El recurso fue editado con exito. Se generó la siguiente URL:\n" + photoURL)
	return c.SendStatus(202)
}

var PutIdHandler = func(c *fiber.Ctx) error {
	carIndex, idInt := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
	responseCode, carToEdit := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckRequestDataAndCarIndex, carIndex)
	if responseCode != 0 {
		return c.SendStatus(responseCode)
	}
	carToEdit.Cylinders = availableCars[carIndex].Cylinders
	carToEdit.FuelType = availableCars[carIndex].FuelType
	carToEdit.PhotoURL = availableCars[carIndex].PhotoURL
	carToEdit.VerifiedURL = availableCars[carIndex].VerifiedURL
	availableCars[carIndex] = carToEdit
	availableCars[carIndex].Id = idInt
	c.SendString("El recurso fue editado con exito.")
	return c.SendStatus(202)

}

var DeleteHandler = func(c *fiber.Ctx) error {
	carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
	responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndex, carIndex)
	if responseCode != 0 {
		return c.SendStatus(responseCode)
	}
	if availableCars[carIndex].PhotoURL != "" {
		fileName := "photo_" + strconv.Itoa(availableCars[carIndex].Id) + "." + appAuxLib.GetFileExtensionFromURL(availableCars[carIndex].PhotoURL)
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
