package main

import (
	"fmt"
	"os"
	"reto/appAuxLib"
	"reto/awsAuxLib"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var availableCars = []appAuxLib.CarSpecs{
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

func main() {
	app := fiber.New()

	awsAuxLib.S3.Region = "us-east-1"
	awsAuxLib.S3.NewSession(awsAuxLib.S3.Region)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(availableCars)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		responseCode, newCar := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckRequestData, 0)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		newCar.Id = appAuxLib.GetNewIntId(availableCars)
		newCar.PhotoURL = ""
		newCar.VerifiedURL = false
		availableCars = append(availableCars, newCar)
		c.SendString("El recurso fue añadido con exito.")
		return c.SendStatus(201)
	})
	app.Get("/export", func(c *fiber.Ctx) error {
		f := appAuxLib.GetANewExcelizeFileOfCarSpecsSlice(availableCars)
		if err := f.SaveAs("./public/temp/BookTemp1.xlsx"); err != nil {
			fmt.Println(err)
		}
		c.SendFile("./public/temp/BookTemp1.xlsx")
		os.Remove("./public/temp/BookTemp1.xlsx")
		return c.SendStatus(202)
	})

	app.Post("/import", func(c *fiber.Ctx) error {
		responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckExcelFile, 0)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		file, _ := c.FormFile("excelFile")
		fileAndPath := "./public/temp/" + file.Filename
		c.SaveFile(file, fileAndPath)
		var err error
		availableCars, err = appAuxLib.ImportDartaFromExcelFile(fileAndPath, availableCars)
		if err != nil {
			return c.SendStatus(400)
		}
		os.Remove("./public/temp/" + file.Filename)
		return c.SendStatus(202)
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
		responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndex, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		return c.Status(fiber.StatusOK).JSON(availableCars[carIndex])
	})
	app.Put("/:id/client-upload", func(c *fiber.Ctx) error {
		carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
		responseCode, _ := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckCarIndex, carIndex)
		if responseCode != 0 {
			return c.SendStatus(responseCode)
		}
		availableCars[carIndex].VerifiedURL = true
		c.SendString("La URL fue verificada con exito.")
		return c.SendStatus(202)

	})
	app.Put("/:id/set-photo/:mode", func(c *fiber.Ctx) error {
		carIndex, _ := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
		if c.Params("mode") == "client-upload" || c.Params("mode") == "server-upload" {
			var photoURL string
			var photoFileExtension string
			if c.Params("mode") == "server-upload" {
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
				availableCars[carIndex].VerifiedURL = true
				os.Remove("./public" + pathAndFile)
			} else {
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
			c.SendString("El recurso fue editado con exito. Se generó la siguiente URL:\n" + photoURL)
			availableCars[carIndex].PhotoURL = strings.Split(photoURL, "?")[0]
		} else {
			c.SendString("Method Not Allowed")
			return c.SendStatus(405)
		}
		return c.SendStatus(202)
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		carIndex, idInt := appAuxLib.GetIndexOfStringId(c.Params("id"), availableCars)
		responseCode, carToEdit := appAuxLib.GetMeAReponseAndOrANewCar(c, appAuxLib.CheckRequestDataAndCarIndex, carIndex)
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

	app.Delete("/:id/remove-photo", deleteHandler)
	app.Delete("/:id", deleteHandler)

	//app.Static("/photos", "./public/photos")
	app.Listen(":3000")
}
