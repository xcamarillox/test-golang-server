package goHelpers

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	IntegerType                = "int"
	Float32Type                = "float32"
	BoolType                   = "boolType"
	StringType                 = "string"
	valueToStringMode          = 1
	getReflectedFieldValueMode = 2
	setReflectedFieldValueMode = 3
)

func CheckIfItsAStruct(myStruct interface{}) bool {
	//fmt.Println(reflect.ValueOf(myStruct).Kind(), reflect.ValueOf(myStruct).Type(), reflect.ValueOf(myStruct).Type().String())
	if reflect.ValueOf(myStruct).Kind() != reflect.Struct {
		return false
	}
	return true
}

func GetStructTotalFields(myStruct interface{}) int {
	if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
		t := reflect.TypeOf(myStruct)
		return t.NumField()
	}
	return -1
}

func GetStructFieldNames(myStruct interface{}) ([]string, error) {
	var names []string
	if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
		t := reflect.TypeOf(myStruct)
		names = make([]string, t.NumField())
		for i := range names {
			names[i] = t.Field(i).Name
		}
		return names, nil
	}
	return names, errors.New("Not a struct!")
}

func GetStructFieldValue(myStruct interface{}, fieldName string) (interface{}, string, error) {
	//if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
	reflectedValue, _ := getReflectedField(&myStruct, fieldName)
	return typeSwitchProcessing(reflectedValue, getReflectedFieldValueMode, "")
	//}
	//return myStruct, "", errors.New("Not a struct!")
}

func SetStructFieldValue(myStruct interface{}, fieldName string, fieldValue string) (interface{}, string, error) {
	//if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
	reflectedValue, err := getReflectedField(myStruct, fieldName)
	value, typeInString, err := typeSwitchProcessing(reflectedValue, setReflectedFieldValueMode, fieldValue)
	return value, typeInString, err
	//}
	//return myStruct, "", errors.New("Not a struct!")
}

func ValueToStringConversion(valueToConvert interface{}, typeInString string) (string, string, error) {
	//if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
	value, typeInString, err := typeSwitchProcessing(valueToConvert, valueToStringMode, typeInString)
	return value.(string), typeInString, err
	//}
	//return myStruct, "", errors.New("Not a struct!")
}

/////////////////PRIVATE FUNCTIONS/////////////////

func getReflectedField(myStruct interface{}, fieldName string) (reflect.Value, error) {
	var curField reflect.Value
	err := errors.New("Not a struct!")
	//if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
	pointToStruct := reflect.ValueOf(myStruct)
	curStruct := pointToStruct.Elem()
	curField = curStruct.FieldByName(fieldName)
	if curField.IsValid() {
		return curField, nil
	}
	err = errors.New(fieldName + " not a found!")
	//}
	return curField, err
}

func typeSwitchProcessing(conversionValue interface{}, conversionMode int, optionString string) (interface{}, string, error) {
	//ValueToStringMode: Args(valueWithSomeType, ValueToStringMode, when optionString="" value its inferred on conversionValue, when optionString="int" for example the value its forced to int)
	//GetReflectedFieldValueMode: Args(reflect.Value, GetReflectedFieldValueMode, optionString = notUsedInThisModeString: "" for example)
	//SetReflectedFieldValueMode: Args(reflect.Value, SetReflectedFieldValueMode, optionString = stringValueToSetToReflectedValue: "1234" for example)
	typeInString := optionString
	if conversionMode == valueToStringMode && optionString == "" {
		typeInString = reflect.ValueOf(conversionValue).Kind().String()
	}
	if conversionMode == getReflectedFieldValueMode || conversionMode == setReflectedFieldValueMode {
		typeInString = conversionValue.(reflect.Value).Kind().String()
	}
	switch typeInString {
	case IntegerType:
		if valueToStringMode == conversionMode {
			value := conversionValue.(int)
			return strconv.Itoa(value), typeInString, nil
		}
		if getReflectedFieldValueMode == conversionMode {
			return int(conversionValue.(reflect.Value).Int()), typeInString, nil
		}
		if setReflectedFieldValueMode == conversionMode {
			value, err := strconv.ParseInt(optionString, 0, 0)
			conversionValue.(reflect.Value).SetInt(value)
			return value, typeInString, err
		}
	case Float32Type:
		if valueToStringMode == conversionMode {
			value := conversionValue.(float32)
			return fmt.Sprintf("%f", value), typeInString, nil
		}
		if getReflectedFieldValueMode == conversionMode {
			return float32(conversionValue.(reflect.Value).Float()), typeInString, nil
		}
		if setReflectedFieldValueMode == conversionMode {
			value, err := strconv.ParseFloat(optionString, 32)
			conversionValue.(reflect.Value).SetFloat(value)
			return value, typeInString, err
		}
	case BoolType:
		if valueToStringMode == conversionMode {
			value := conversionValue.(bool)
			return fmt.Sprintf("%t", value), typeInString, nil
		}
		if getReflectedFieldValueMode == conversionMode {
			return conversionValue.(reflect.Value).Bool(), typeInString, nil
		}
		if setReflectedFieldValueMode == conversionMode {
			value, err := strconv.ParseBool(optionString)
			conversionValue.(reflect.Value).SetBool(value)
			return value, typeInString, err
		}
	case StringType:
		if valueToStringMode == conversionMode {
			return conversionValue.(string), typeInString, nil
		}
		if getReflectedFieldValueMode == conversionMode {
			return conversionValue.(reflect.Value).String(), typeInString, nil
		}
		if setReflectedFieldValueMode == conversionMode {
			conversionValue.(reflect.Value).SetString(optionString)
			return optionString, typeInString, nil
		}
	default:
		fmt.Println("Type is unknown or not implemented!")
	}
	return "", "", errors.New("Type is unknown or not implemented!")
}
