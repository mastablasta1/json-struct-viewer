package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/labstack/echo/v4"
)

/*
Объявите переменные
*/
var AppPort = "8080"
var FilePath string = "/home/esk/work/src/json-struct-viewer/json_files/"

type FieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Response struct {
	Structure map[string]interface{}   `json:"structure"`
	Records   []map[string]interface{} `json:"records"`
}

// рекурсивная функция для получения структуры данных
func getStructure(data interface{}) map[string]interface{} {
	structure := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}: // Если это объект
		for key, value := range v {
			structure[key] = getStructure(value) // Рекурсивный вызов
		}
	case []interface{}: // Если это массив
		if len(v) > 0 {
			structure["array"] = getStructure(v[0]) // Предполагаем, что все элементы массива имеют одинаковую структуру
		}
	default:
		structure["type"] = reflect.TypeOf(data).String() // Указываем тип данных
	}

	return structure
}

func getStructAndRecordsFromJSON(filePath string) (Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Response{}, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return Response{}, err
	}

	var data interface{}
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return Response{}, err
	}

	structFields := getStructure(data) // Получаем структуру данных
	var records []map[string]interface{}

	// Проверяем тип данных
	switch v := data.(type) {
	case []interface{}: // Если это массив
		// Получаем первые 4 записи
		for i := 0; i < len(v) && i < 4; i++ {
			if record, ok := v[i].(map[string]interface{}); ok {
				records = append(records, record)
			}
		}
	case map[string]interface{}: // Если это объект
		records = append(records, v) // Добавляем сам объект как единственную запись
	default:
		return Response{}, fmt.Errorf("unsupported JSON structure")
	}

	return Response{Structure: structFields, Records: records}, nil
}

func getStructHandler(c echo.Context) error {
	fileName := c.QueryParam("file")
	if fileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file parameter is required"})
	}

	filePath := FilePath + fileName // Укажите путь к папке с JSON файлами
	response, err := getStructAndRecordsFromJSON(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func main() {
	e := echo.New()

	e.GET("/get-struct", getStructHandler)

	e.Logger.Fatal(e.Start(":" + AppPort))
}
