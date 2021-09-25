package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	_ "reflect"
	"strconv"
	"strings"
	"time"
)

const (
	EnvName = "ENV FILE"
)

type CompanyList struct {
	Companies []InInfoCompany
}
type InInfoCompany struct {
	Company   string      `json:"company"`
	Type      string      `json:"type"`
	Value     int         `json:"value"`
	Id        interface{} `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	Valid     bool        `default:"true"`
}

type OutInfoCompany struct {
	Company             string        `json:"company"`
	ValidCountOperation int           `json:"valid_operations_count"`
	Balance             int           `json:"balance"`
	InvalidOperation    []interface{} `json:"invalid_operations"`
}

func readFileName() string {
	var name *string
	var nameEnv string

	name = flag.String("file", "", "File to open for read")
	flag.Parse()
	if *name != "" {
		return *name
	}

	nameEnv, ok := os.LookupEnv(EnvName)
	if ok == true {
		return nameEnv
	}

	fmt.Println("Input file name:")
	fmt.Scanf("%s", name)
	return *name
}

func UnmarshalValue(value interface{}) (int, bool) {
	var answer int
	var bValue bool
	switch value.(type) {
	case string:
		tmp, err := strconv.Atoi(value.(string))
		bValue = true
		if err != nil {
			bValue = false
		}
		answer = tmp
	case float64:
		answer = int(value.(float64))
		bValue = true
	case float32:
		answer = int(value.(float32))
		bValue = true
	case int:
		answer = value.(int)
		bValue = true
	default:
		fmt.Println("не получилось value")
	}
	return answer, bValue
}

func UnmarshalCompany(key interface{}) (string, bool) {
	var companyName string
	var bCompany bool
	switch key.(type) {
	case string:
		companyName = key.(string)
		bCompany = true
	default:
		bCompany = false
	}
	return companyName, bCompany
}

func UnmarshalType(key interface{}) (string, bool) {
	var typ string
	var bType bool
	typ, ok := key.(string)
	if ok {
		typ = strings.ToLower(typ)
		if typ == "-" || typ == "+" || typ == "outcome" || typ == "income" {
			bType = true
		}
	}
	return typ, bType
}

func UnmarshalId(key interface{}) (interface{}, bool) {
	var id interface{}
	var tId bool
	switch key.(type) {
	case int:
		id = key.(int)
		tId = true
	case float64:
		id = int(key.(float64))
		tId = true
	case string:
		id = key.(string)
		tId = true
	}
	return id, tId
}

func UnmarshalDate(key interface{}) (time.Time, bool) {
	var t time.Time
	var bTime bool
	var err error
	tStr, ok := key.(string)
	if ok {
		t, err = time.Parse(time.RFC3339, tStr)
		bTime = true
		if err != nil {
			bTime = false
		}
	}
	return t, bTime
}

// func (i *InInfoCompany) UnmarshalOperation(objMap map[string]interface{}) error {
// 	for key, value := range objMap {
// 		switch key {
// 		case "company":
// 			i.Company, bCompany = UnmarshalCompany(value)
// 		case "type":
// 			i.Type, bType = UnmarshalType(value)
// 		case "value":
// 			i.Value, bValue = UnmarshalValue(value)
// 		case "id":
// 			i.Id, bId = UnmarshalId(value)
// 			switch i.Id.(type) {
// 			case int:
// 				i.Id = i.Id.(int)
// 			case string:
// 				i.Id = i.Id.(string)
// 			}
// 		case "created_at":
// 			i.CreatedAt, bTime = UnmarshalDate(value)
// 		}
// 	}
// }

func (i *InInfoCompany) MyUnmarshalObj(objMap map[string]interface{}) error {
	var bType bool
	var bValue bool
	var bId bool
	var bTime bool
	var bCompany bool

	for key, value := range objMap {
		switch key {
		case "company":
			i.Company, bCompany = UnmarshalCompany(value)
		case "type":
			i.Type, bType = UnmarshalType(value)
		case "value":
			i.Value, bValue = UnmarshalValue(value)
		case "id":
			i.Id, bId = UnmarshalId(value)
			switch i.Id.(type) {
			case int:
				i.Id = i.Id.(int)
			case string:
				i.Id = i.Id.(string)
			}
		case "created_at":
			i.CreatedAt, bTime = UnmarshalDate(value)
		case "operation":

		}
	}
	if bCompany && bType && bValue && bTime && bId {
		i.Valid = true
	}
	return nil
}

func (i *CompanyList) UnmarshalJSON(data []byte) error {
	var objMapSlice []map[string]interface{}

	if err := json.Unmarshal(data, &objMapSlice); err != nil {
		fmt.Printf("%s", err)
		return errors.New("Can't unmarshal")
	}
	for _, objMap := range objMapSlice {
		var companyInfo InInfoCompany
		_ = companyInfo.MyUnmarshalObj(objMap)
		i.Companies = append(i.Companies, companyInfo)
	}
	// for _, value := range i.Companies {
	// 	fmt.Printf("%+v\n", value)
	// }
	fmt.Println("Я тут")

	return nil
}

func main() {
	filename := readFileName()
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Can't to open file:", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	var test CompanyList
	_ = test.UnmarshalJSON(data)

}
