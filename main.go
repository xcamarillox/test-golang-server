package main

import (
	"reto/appAuxLib"
	"reto/awsAuxLib"
	"reto/reqHandlers"

	"github.com/gofiber/fiber/v2"
)

func main() {

	reqHandlers.SetTestData([]appAuxLib.CarSpecs{
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
	})

	app := fiber.New()

	awsAuxLib.S3.Region = "us-east-1"
	awsAuxLib.S3.NewSession(awsAuxLib.S3.Region)

	app.Get("/", reqHandlers.GetRootHandler)
	app.Post("/", reqHandlers.PostRootHandler)
	app.Get("/export", reqHandlers.GetExportHandler)
	app.Post("/import", reqHandlers.PostImportHandler)
	app.Get("/:id", reqHandlers.GetIdHandler)
	app.Put("/:id/client-upload", reqHandlers.PutIdClientUploadHandler)
	app.Put("/:id/set-photo/:mode", reqHandlers.PutIdSetPhotoModeHandler)
	app.Put("/:id", reqHandlers.PutIdHandler)
	app.Delete("/:id/remove-photo", reqHandlers.DeleteHandler)
	app.Delete("/:id", reqHandlers.DeleteHandler)
	//app.Static("/photos", "./public/photos")
	app.Listen(":3000")
}
