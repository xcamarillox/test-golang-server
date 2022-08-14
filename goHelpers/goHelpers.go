package goHelpers

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	ValueToStringMode          = 1
	GetReflectedFieldValueMode = 2
	SetReflectedFieldValueMode = 3
	ReflectToStringMode        = 4
	integerType                = "int"
	float32Type                = "float32"
	boolType                   = "boolType"
	stringType                 = "string"
)

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

func CheckIfItsAStruct(myStruct interface{}) bool {
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
	reflectedValue, err := getReflectedField(myStruct, fieldName)
	value, typeString, err := MakeAConversion(reflectedValue, GetReflectedFieldValueMode, "")
	return value, typeString, err
	//}
	//return myStruct, "", errors.New("Not a struct!")
}

func SetStructFieldValue(myStruct interface{}, fieldName string, fieldValue string) (interface{}, string, error) {
	//if isAStruct := CheckIfItsAStruct(myStruct); isAStruct {
	reflectedValue, err := getReflectedField(myStruct, fieldName)
	value, typeString, err := MakeAConversion(reflectedValue, SetReflectedFieldValueMode, fieldValue)
	return value, typeString, err
	//}
	//return myStruct, "", errors.New("Not a struct!")
}

func MakeAConversion(conversionValue interface{}, conversionMode int, optionString string) (interface{}, string, error) {
	//ValueToStringMode: Args(valueWithSomeType, ValueToStringMode, when optionString="" value its inferred on conversionValue, when optionString="int" for example the value its forced to int)
	//GetReflectedFieldValueMode: Args(reflect.Value, GetReflectedFieldValueMode, optionString = notUsedInThisModeString: "" for example)
	//SetReflectedFieldValueMode: Args(reflect.Value, SetReflectedFieldValueMode, optionString = stringValueToSetToReflectedValue: "1234" for example)
	typeString := optionString
	if conversionMode == ValueToStringMode && optionString == "" {
		typeString = reflect.ValueOf(conversionValue).Kind().String()
	}
	if conversionMode == ReflectToStringMode && optionString == "" {
		//typeString = reflect.ValueOf(conversionValue).Kind().String()
	}
	if conversionMode == GetReflectedFieldValueMode || conversionMode == SetReflectedFieldValueMode {
		typeString = conversionValue.(reflect.Value).Kind().String()
	}
	switch typeString {
	case integerType:
		if ValueToStringMode == conversionMode {
			value := conversionValue.(int)
			return strconv.Itoa(value), typeString, nil
		}
		if ReflectToStringMode == conversionMode {
			value := conversionValue.(reflect.Value).Int()
			return strconv.Itoa(int(value)), typeString, nil
		}
		if GetReflectedFieldValueMode == conversionMode {
			return conversionValue, typeString, nil
		}
		if SetReflectedFieldValueMode == conversionMode {
			value, err := strconv.ParseInt(optionString, 0, 0)
			conversionValue.(reflect.Value).SetInt(value)
			return value, typeString, err
		}
	case float32Type:
		if ValueToStringMode == conversionMode {
			value := conversionValue.(float32)
			return fmt.Sprintf("%f", value), typeString, nil
		}
		if ReflectToStringMode == conversionMode {
			value := conversionValue.(reflect.Value).Float()
			return fmt.Sprintf("%f", value), typeString, nil
		}
		if GetReflectedFieldValueMode == conversionMode {
			return conversionValue, typeString, nil
		}
		if SetReflectedFieldValueMode == conversionMode {
			value, err := strconv.ParseFloat(optionString, 32)
			conversionValue.(reflect.Value).SetFloat(value)
			return value, typeString, err
		}
	case boolType:
		if ValueToStringMode == conversionMode {
			value := conversionValue.(bool)
			return fmt.Sprintf("%t", value), typeString, nil
		}
		if ReflectToStringMode == conversionMode {
			value := conversionValue.(reflect.Value).Bool()
			return fmt.Sprintf("%t", value), typeString, nil
		}
		if GetReflectedFieldValueMode == conversionMode {
			return conversionValue, typeString, nil
		}
		if SetReflectedFieldValueMode == conversionMode {
			value, err := strconv.ParseBool(optionString)
			conversionValue.(reflect.Value).SetBool(value)
			return value, typeString, err
		}
	case stringType:
		if ValueToStringMode == conversionMode {
			return conversionValue.(string), typeString, nil
		}
		if ReflectToStringMode == conversionMode {
			return conversionValue.(reflect.Value).String(), typeString, nil
		}
		if GetReflectedFieldValueMode == conversionMode {
			return conversionValue, typeString, nil
		}
		if SetReflectedFieldValueMode == conversionMode {
			conversionValue.(reflect.Value).SetString(optionString)
			return optionString, typeString, nil
		}
	default:
		fmt.Println("Type is unknown or not implemented!")
	}
	return "", "", errors.New("Type is unknown or not implemented!")
}
