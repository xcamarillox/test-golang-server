package appAuxLib

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
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
				c.SendString("Debes anexar el campo photoFile además de un archivo de imagen válido. Los tipos aceptados son jpg, jpeg, png o gif exclusivamente.")
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
	if mode == CheckExcelFile {
		file, err := c.FormFile("excelFile")
		if err != nil {
			c.SendString("Debes anexar el campo excelFile además de un archivo de Excel con extensión xlsx.")
			return 404, newCar
		}
		splitFileName := strings.Split(file.Filename, ".")
		extension := strings.ToLower(splitFileName[len(splitFileName)-1])
		if extension != "xlsx" || len(splitFileName) < 2 {
			if len(splitFileName) < 2 {
				c.SendString("Error con el nombre de archivo. El archivo debe tener nombre y extensión.")
			} else {
				c.SendString("Error en el tipo de archivo. Solo es aceptada la extensión xlsx.")
			}
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
	} else {
		if index == 0 {
			index = 26
		}
		return string(l[index-1]), nil
	}

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

func GetCarSpecsFieldsNames() []string {
	t := reflect.TypeOf(CarSpecs{})
	names := make([]string, t.NumField())
	for i := range names {
		names[i] = t.Field(i).Name
	}
	return names
}

func GetOrSetReflectedFieldValue(structValue reflect.Value, isSetMode bool, stringValue string) (interface{}, error) {
	//fmt.Println(structValue.Kind().String())
	switch t := structValue.Kind().String(); t {
	case "int":
		if isSetMode {
			value, err := strconv.ParseInt(stringValue, 0, 0)
			structValue.SetInt(value)
			return nil, err
		} else {
			integerValue, _ := structValue.Interface().(int)
			return integerValue, nil
		}
	case "float32":
		if isSetMode {
			value, err := strconv.ParseFloat(stringValue, 32)
			structValue.SetFloat(value)
			return nil, err
		} else {
			floatValue, _ := structValue.Interface().(float32)
			return floatValue, nil
		}
	case "string":
		if isSetMode {
			structValue.SetString(stringValue)
			return nil, nil
		} else {
			stringValue, _ := structValue.Interface().(string)
			return stringValue, nil
		}
	case "bool":
		if isSetMode {
			value, err := strconv.ParseBool(stringValue)
			structValue.SetBool(value)
			return nil, err
		} else {
			boolValue, _ := structValue.Interface().(bool)
			return boolValue, nil
		}
	default:
		fmt.Println("Type is unknown!")
		return "", errors.New("El tipo de dato es desconocido.")
	}
}

func GetReflectField(obj interface{}, fieldName string) reflect.Value {
	pointToStruct := reflect.ValueOf(obj) // addressable
	curStruct := pointToStruct.Elem()
	if curStruct.Kind() != reflect.Struct {
		fmt.Println("not struct")
	}
	curField := curStruct.FieldByName(fieldName) // type: reflect.Value
	if !curField.IsValid() {
		fmt.Println("not found:" + fieldName)
	}
	return curField
}

func GetANewExcelizeFileOfCarSpecsSlice(availableCars []CarSpecs) *excelize.File {
	fieldsNames := GetCarSpecsFieldsNames()
	f := excelize.NewFile()
	for j := range fieldsNames {
		columnPosition, _ := IndexToColumn(j + 1)
		cellPosition := columnPosition + strconv.Itoa(1)
		f.SetCellValue("Sheet1", cellPosition, fieldsNames[j])
	}
	for i := range availableCars {
		for j := range fieldsNames {
			columnPosition, _ := IndexToColumn(j + 1)
			cellPosition := columnPosition + strconv.Itoa(2+i)
			cellValue, _ := GetOrSetReflectedFieldValue(GetReflectField(&availableCars[i], fieldsNames[j]), false, "")
			f.SetCellValue("Sheet1", cellPosition, cellValue)
		}
	}
	return f
}

func ImportDataFromExcelFile(filePath string, availableCars []CarSpecs) ([]CarSpecs, []string, error) {
	var rowErr []string
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return []CarSpecs{}, rowErr, err
	}
	//Acá se revisa si la primer file corresponde a los campos(fields) del struct de datos
	fieldsNames := GetCarSpecsFieldsNames()
	for j := range fieldsNames {
		columnPosition, _ := IndexToColumn(j + 1)
		cellPosition := columnPosition + strconv.Itoa(1)
		cellValue, err := f.GetCellValue("Sheet1", cellPosition)
		if err != nil {
			return []CarSpecs{}, rowErr, err
		}
		if cellValue != fieldsNames[j] {
			return []CarSpecs{}, rowErr, errors.New("la estructura del excel es incorrecta")
		}
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return []CarSpecs{}, rowErr, err
	}
	/*
		Acá se revisa si el valor de la primer celda de la fila corresponde a un ID,
		si es así los datos de la fila se reflejan en los registros. En caso contario
		Se crea un nuevo ID con los datos segun correspondan.
	*/
	rows = append(rows[1:])
	//fmt.Println("--------------------------")
	//fmt.Println("Rows:", rows)
	//ROWS:
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}
		idxId, _ := GetIndexOfStringId(row[0], availableCars)
		newCar := CarSpecs{}
		for j := range fieldsNames {
			if j == 0 || j == 7 || j == 8 { //ID, PhotoURL, VerifiedURL
				continue
			}
			columnPosition, _ := IndexToColumn(j + 1)
			cellPosition := columnPosition + strconv.Itoa(i+2)
			cellValue, err := f.GetCellValue("Sheet1", cellPosition)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(cellPosition, cellValue)
			//fmt.Println(newCar, fieldsNames[j])
			_, convErr := GetOrSetReflectedFieldValue(GetReflectField(&newCar, fieldsNames[j]), true, cellValue)
			if convErr != nil || newCar.Year < 0 {
				convErr = errors.New("Nuevo error")
				rowErr = append(rowErr, cellPosition)
				//fmt.Println("Error en la fila: " + strconv.Itoa(i+2) + ". Celda: " + cellPosition)
				//fmt.Println("Error en tipo de dato con: " + cellValue)
				//fmt.Println()
				//continue ROWS
			}
		}
		//fmt.Println(newCar)
		if idxId < 0 {
			newCar.Id = GetNewIntId(availableCars)
			newCar.PhotoURL = ""
			newCar.VerifiedURL = false
			availableCars = append(availableCars, newCar)
		} else {
			newCar.Id = availableCars[idxId].Id
			newCar.PhotoURL = availableCars[idxId].PhotoURL
			newCar.VerifiedURL = availableCars[idxId].VerifiedURL
			availableCars[idxId] = newCar
		}
	}
	//fmt.Println(availableCars)
	return availableCars, rowErr, nil
}
