package appAuxLib

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reto/awsAuxLib"
	"reto/goHelpers"

	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

const (
	CheckCarIndex                      = 1
	CheckRequestData                   = 2
	CheckCarIndexAndImage              = 3
	CheckRequestDataAndCarIndex        = 4
	CheckCarIndexAndPhotoFileNameField = 5
	CheckExcelFile                     = 6
)

type CarSpecs struct {
	Id               int     `json:"id"`
	Make             string  `json:"make"`
	Model            string  `json:"model"`
	Year             int     `json:"year"`
	EngineCapacity   float32 `json:"engineCapacity"` //enlitros
	Color            string  `json:"color"`
	TransmissionType string  `json:"transmissionType"` // manual / automatica
	Cylinders        int     `json:"cylinders"`
	FuelType         string  `json:"fuelType"`
	PhotoURL         string  `json:"photoURL"`
	VerifiedURL      bool    `json:"verifiedURL"`
}

type ErrorMessageAPI struct {
	Message string `json:"message"`
}

type CarAPIInfo struct {
	Model     string `json:"model"`
	Year      int    `json:"year"`
	Cylinders int    `json:"cylinders"`
	FuelType  string `json:"fuel_type"`
}

func GetIndexOfIntId(id int, availableCars []CarSpecs) int {
	for index, car := range availableCars {
		if car.Id == id {
			return index
		}
	}
	return -1
}

func GetIndexOfStringId(id string, availableCars []CarSpecs) (int, int) { // returns carIndex, idInt
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
			return "", "Error con el nombre de archivo. El archivo debe tener nombre y extensi??n."
		}
		return "", "Error en el tipo de archivo. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente."
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
			c.SendString("El a??o de manufactura del autom??vil no puede ser negativo.")
			return 400, newCar
		}
		if newCar.TransmissionType != "automatica" && newCar.TransmissionType != "manual" {
			c.SendString("TransmissionType solo puede ser automatica o manual")
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
				c.SendString("Debes anexar el campo photoFile adem??s de un archivo de imagen v??lido. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
				return 404, newCar
			}
			_, errorString = GetPhotoFileExtension(file.Filename)
		}
		if mode == CheckCarIndexAndPhotoFileNameField {
			field := c.FormValue("photoFileName")
			if field == "" {
				c.SendString("Debes anexar el campo photoFileName adem??s del nombre y extensi??n de tu im??gen. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
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
	if mode == CheckExcelFile {
		file, err := c.FormFile("excelFile")
		if err != nil {
			c.SendString("Debes anexar el campo excelFile adem??s de un archivo de Excel con extensi??n xlsx.")
			return 404, newCar
		}
		splitFileName := strings.Split(file.Filename, ".")
		extension := strings.ToLower(splitFileName[len(splitFileName)-1])
		if extension != "xlsx" || len(splitFileName) < 2 {
			if len(splitFileName) < 2 {
				c.SendString("Error con el nombre de archivo. El archivo debe tener nombre y extensi??n.")
			}
			c.SendString("Error en el tipo de archivo. Solo es aceptada la extensi??n xlsx.")
			return 400, newCar
		}
	}
	return 0, newCar //incoming data is ok and setted in car, 0 returned as no error code
}

// indexToColumn takes in an index value & converts it to A1 Notation
// Index 1 is Column A
// E.g. 3 == C, 29 == AC, 731 == ABC
func IndexToColumn(index int) (string, error) {

	// Validate index size
	maxIndex := 18278
	if index > maxIndex {
		return "", fmt.Errorf("index cannot be greater than %v (column ZZZ)", maxIndex)
	}

	// Get column from index
	l := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if index > 26 {
		letterA, _ := IndexToColumn(int(math.Floor(float64(index-1) / 26)))
		letterB, _ := IndexToColumn(index % 26)
		return letterA + letterB, nil
	}
	if index == 0 {
		index = 26
	}
	return string(l[index-1]), nil
}

// columnToIndex takes in A1 Notation & converts it to an index value
// Column A is index 1
// E.g. C == 3, AC == 29, ABC == 731
func ColumnToIndex(column string) (int, error) {
	// Calculate index from column string
	var index int
	var a uint8 = "A"[0]
	var z uint8 = "Z"[0]
	var alphabet = z - a + 1
	i := 1
	for n := len(column) - 1; n >= 0; n-- {
		r := column[n]
		if r < a || r > z {
			return 0, fmt.Errorf("invalid character in column, expected A-Z but got [%c]", r)
		}
		runePos := int(r-a) + 1
		index += runePos * int(math.Pow(float64(alphabet), float64(i-1)))
		i++
	}
	// Return column index & success
	return index, nil
}

func GetANewExcelizeFileOfCarSpecsSlice(availableCars []CarSpecs) *excelize.File {
	fieldsNames, _ := goHelpers.GetStructFieldNames(CarSpecs{})
	f := excelize.NewFile()
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: "#000000",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	for j := range fieldsNames {
		columnPosition, _ := IndexToColumn(j + 1)
		cellPosition := columnPosition + strconv.Itoa(1)
		f.SetCellStyle("Sheet1", cellPosition, cellPosition, style)
		f.SetCellValue("Sheet1", cellPosition, fieldsNames[j])
	}
	for i := range availableCars {
		for j := range fieldsNames {
			columnPosition, _ := IndexToColumn(j + 1)
			cellPosition := columnPosition + strconv.Itoa(2+i)
			cellValue, _, _ := goHelpers.GetStructFieldValue(&availableCars[i], fieldsNames[j])
			f.SetCellStyle("Sheet1", cellPosition, cellPosition, style)
			f.SetCellValue("Sheet1", cellPosition, cellValue)
		}
	}
	dvRange := excelize.NewDataValidation(true)
	dvRange.Sqref = "G2:G100"
	dvRange.SetDropList([]string{"manual", "automatica"})
	f.AddDataValidation("Sheet1", dvRange)
	return f
}

func ImportDataFromExcelFile(filePath string, availableCars []CarSpecs) ([]CarSpecs, []string, error) {
	var cellsWithErr []string
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return []CarSpecs{}, cellsWithErr, err
	}
	//Ac?? se revisa si la primer file corresponde a los campos(fields) del struct de datos
	fieldsNames, _ := goHelpers.GetStructFieldNames(CarSpecs{})
	for j := range fieldsNames {
		columnPosition, _ := IndexToColumn(j + 1)
		cellPosition := columnPosition + strconv.Itoa(1)
		cellValue, err := f.GetCellValue("Sheet1", cellPosition)
		if err != nil {
			return []CarSpecs{}, cellsWithErr, err
		}
		if cellValue != fieldsNames[j] {
			return []CarSpecs{}, cellsWithErr, errors.New("la estructura del excel es incorrecta")
		}
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return []CarSpecs{}, cellsWithErr, err
	}
	/*
		Ac?? se revisa si el valor de la primer celda de la fila corresponde a un ID,
		si es as?? los datos de la fila se reflejan en los registros. En caso contario
		Se crea un nuevo ID con los datos segun correspondan.
	*/
	rows = append(rows[1:])
	//fmt.Println("Rows:", rows)
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: "#000000",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	//ROWS:
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}
		idxId, _ := GetIndexOfStringId(row[0], availableCars)
		newCar := CarSpecs{}
		for j := range fieldsNames {
			errFlags := false
			if j == 0 || j == 7 || j == 8 || j == 9 || j == 10 { //ID, PhotoURL, VerifiedURL, Cylinders, FuelType
				continue
			}
			columnPosition, _ := IndexToColumn(j + 1)
			cellPosition := columnPosition + strconv.Itoa(i+2)
			err = f.SetCellStyle("Sheet1", cellPosition, cellPosition, style)
			cellValue, err := f.GetCellValue("Sheet1", cellPosition)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(newCar, fieldsNames[j], "set")
			_, _, convErr := goHelpers.SetStructFieldValue(&newCar, fieldsNames[j], cellValue)
			//_, _, convErr := goHelpers.SetStructFieldValue(newCar, fieldsNames[j], cellValue)
			if convErr != nil {
				//convErr = errors.New("Nuevo error")
				errFlags = true
				cellsWithErr = append(cellsWithErr, cellPosition)
				//continue ROWS
			}
			if fieldsNames[j] == "Year" && newCar.Year < 0 {
				errFlags = true
				newCar.Year = 0
				cellsWithErr = append(cellsWithErr, cellPosition)
			}
			if fieldsNames[j] == "TransmissionType" && newCar.TransmissionType != "automatica" && newCar.TransmissionType != "manual" {
				errFlags = true
				newCar.TransmissionType = "manual"
				cellsWithErr = append(cellsWithErr, cellPosition)
			}
			if errFlags == true && idxId > -1 {
				//fmt.Println(availableCars[idxId], idxId)
				cellValue, cellType, _ := goHelpers.GetStructFieldValue(&availableCars[idxId], fieldsNames[j])
				//cellValue, cellType, _ := goHelpers.GetStructFieldValue(availableCars[idxId], fieldsNames[j])
				cellStringValue, _, _ := goHelpers.ValueToStringConversion(cellValue, cellType)
				goHelpers.SetStructFieldValue(&newCar, fieldsNames[j], cellStringValue)
				//goHelpers.SetStructFieldValue(newCar, fieldsNames[j], cellStringValue)
			}
		}
		//fmt.Println(newCar)
		if idxId < 0 {
			newCar.Id = GetNewIntId(availableCars)
			newCar.PhotoURL = ""
			newCar.VerifiedURL = false
			newCar.Cylinders = 0
			newCar.FuelType = ""
			availableCars = append(availableCars, newCar)
			continue
		}
		newCar.Id = availableCars[idxId].Id
		newCar.PhotoURL = availableCars[idxId].PhotoURL
		newCar.VerifiedURL = availableCars[idxId].VerifiedURL
		newCar.Cylinders = availableCars[idxId].Cylinders
		newCar.FuelType = availableCars[idxId].FuelType
		availableCars[idxId] = newCar
	}
	//fmt.Println(availableCars)
	return availableCars, cellsWithErr, nil
}

func GetURLFileWithMarkedErrors(filePath string, cellsWithErr []string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println(err)
	}
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: "#FF0000",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#FF0000", Style: 4},
			{Type: "top", Color: "#FF0000", Style: 4},
			{Type: "bottom", Color: "#FF0000", Style: 4},
			{Type: "right", Color: "#FF0000", Style: 4},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	for j := range cellsWithErr {
		err = f.SetCellStyle("Sheet1", cellsWithErr[j], cellsWithErr[j], style)
	}
	dvRange := excelize.NewDataValidation(true)
	dvRange.Sqref = "G2:G100"
	dvRange.SetDropList([]string{"manual", "automatica"})
	f.AddDataValidation("Sheet1", dvRange)
	err = f.SaveAs("./public/temp/" + "Errors_1" + ".xlsx")
	defer os.Remove("./public/temp/" + "Errors_1" + ".xlsx")
	if err != nil {
		fmt.Println(err)
	}
	awsAuxLib.S3.UploadObject("./public/temp/"+"Errors_1"+".xlsx", "levita-uploads-dev", "Errors_1"+".xlsx")
	UrlOfFile, _ := awsAuxLib.S3.GetTemporalUrl("levita-uploads-dev", "Errors_1"+".xlsx")
	//fmt.Println(UrlOfFile)
	return UrlOfFile, nil
}

func GetDataAPI(model string, year int) []CarAPIInfo {
	godotenv.Load()
	var carAPIInfo []CarAPIInfo
	var errorMessageAPI ErrorMessageAPI
	client := req.C().
		SetUserAgent("my-custom-client"). // Chainable client settings.
		SetTimeout(5 * time.Second)
	client.R().
		SetHeader("X-Api-Key", os.Getenv("API_KEY_API_NINJAS")). // Chainable request settings.
		SetPathParam("model", model).
		SetPathParam("year", strconv.Itoa(year)).
		SetResult(&carAPIInfo).     // Unmarshal response body into userInfo automatically if status code is between 200 and 299.
		SetError(&errorMessageAPI). // Unmarshal response body into errMsg automatically if status code >= 400.
		EnableDump().               // Enable dump at request level, only print dump content if there is an error or some unknown situation occurs to help troubleshoot.
		Get("https://api.api-ninjas.com/v1/cars?model={model}&year={year}")
	return carAPIInfo
}

func GetCarSpecsWithAPIData(availableCars []CarSpecs) []CarSpecs {
	for index, car := range availableCars {
		if car.Model != "" && car.Year > -1 && (car.Cylinders == 0 || car.FuelType == "") {
			fmt.Println("Solicitud en index:", index)
			carInfo := GetDataAPI(car.Model, car.Year)
			if len(carInfo) > 0 {
				availableCars[index].Cylinders = carInfo[0].Cylinders
				availableCars[index].FuelType = carInfo[0].FuelType
			}
		}
	}
	fmt.Println()
	return availableCars
}
