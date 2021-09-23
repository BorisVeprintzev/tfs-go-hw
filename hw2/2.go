package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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
		if err != nil {
			fmt.Println("Пока так")
		}
		answer = tmp
	case float64:
		answer = int(value.(float64))
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
	switch key.(type) {
	case string:
		tmp := key.(string)
		tmp = strings.ToLower(tmp)
		if tmp == "-" || tmp == "+" || tmp == "outcome" || tmp == "income" {
			typ = tmp
			bType = true
		}
	}
	return typ, bType
}

func UnmarshalId(key interface{}) (interface{}, bool) {
	var id interface{}
	var tId bool
	switch key.(type) {
	case int, string:
		id = key
		tId = true
	}
	return id, tId
}

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
		case "created_at":
		case "operation":

		}
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

		fmt.Println(objMap)
	}
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

	// decoder := json.NewDecoder(file)
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	var test CompanyList
	_ = test.UnmarshalJSON(data)
	// var companyList []InInfoCompany
	// for {
	// 	var company InInfoCompany
	// 	if err := decoder.Decode(&company); err != io.EOF {
	// 		break
	// 	} else if err != nil {
	// 		log.Fatal("Can't decode.", err)
	// 	}
	// 	fmt.Println(company.Company)
	// 	companyList = append(companyList, company)
	// }
	// fmt.Println(companyList)
}
