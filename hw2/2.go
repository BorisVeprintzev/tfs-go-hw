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

type CompanyListIn struct {
	Companies []InInfoCompany `json:"-"`
}

type CompanyMapOut struct {
	CompanySet map[string]*OutInfoCompany
}

type InInfoCompany struct {
	Company   string      `json:"company"`
	Type      string      `json:"type"`
	Value     int         `json:"value"`
	Id        interface{} `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	Valid     bool
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

func (i *InInfoCompany) UnmarshalOperation(objMap map[string]interface{}) (bool, bool, bool, bool, bool) {
	var obType, obValue, obId, obTime, obCompany bool

	for key, value := range objMap {
		switch key {
		case "company":
			i.Company, obCompany = UnmarshalCompany(value)
		case "type":
			i.Type, obType = UnmarshalType(value)
		case "value":
			i.Value, obValue = UnmarshalValue(value)
		case "id":
			i.Id, obId = UnmarshalId(value)
			switch i.Id.(type) {
			case int:
				i.Id = i.Id.(int)
			case string:
				i.Id = i.Id.(string)
			}
		case "created_at":
			i.CreatedAt, obTime = UnmarshalDate(value)
		}
	}
	return obType, obValue, obId, obTime, obCompany
}

func (i *InInfoCompany) MyUnmarshalObj(objMap map[string]interface{}) error {
	var bType, bValue, bId, bTime, bCompany bool
	var obType, obValue, obId, obTime, obCompany bool

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
			tmp, ok := value.(map[string]interface{})
			if ok {
				obType, obValue, obId, obTime, obCompany = i.UnmarshalOperation(tmp)
			}
		}
	}
	if (bCompany || obCompany) && (bType || obType) && (bValue || obValue) && (bTime || obTime) && (bId || obId) {
		i.Valid = true
	}
	return nil
}

func (i *CompanyListIn) UnmarshalJSON(data []byte) error {
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

	return nil
}

func (o *CompanyMapOut) AddNewCompany(info InInfoCompany) error {
	var outInfoCompany OutInfoCompany

	outInfoCompany.Company = info.Company
	if info.Valid == true {
		outInfoCompany.ValidCountOperation += 1
		if info.Type == "outcome" || info.Type == "-" {
			outInfoCompany.Balance -= info.Value
		} else {
			outInfoCompany.Balance += info.Value
		}
	} else {
		outInfoCompany.InvalidOperation = append(outInfoCompany.InvalidOperation, info.Id)
	}
	o.CompanySet[info.Company] = &outInfoCompany
	return nil
}

func NewCompanyListOut(listIn CompanyListIn) CompanyMapOut {
	var companyMapOut CompanyMapOut
	companyMapOut.CompanySet = make(map[string]*OutInfoCompany)

	for _, value := range listIn.Companies {
		var in bool

		for name, info := range companyMapOut.CompanySet {
			if name == value.Company {
				in = true
				if value.Valid == true {
					info.ValidCountOperation += 1
					if value.Type == "income" || value.Type == "+" {
						info.Balance += value.Value
					} else {
						info.Balance -= value.Value
					}
				} else {
					if value.Id != nil {
						info.InvalidOperation = append(info.InvalidOperation, value.Id)
					}
				}
				break
			} else {
				continue
			}
		}
		if in == false {
			companyMapOut.AddNewCompany(value)
		}
	}
	return companyMapOut
}

func main() {
	filename := readFileName()
	fileIn, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fileIn.Close()

	data, err := io.ReadAll(fileIn)
	if err != nil {
		log.Fatal(err)
	}
	var companyListIn CompanyListIn
	_ = companyListIn.UnmarshalJSON(data)

	fileOut, err := os.Create("out.json")
	if err != nil {
		log.Fatal(err)
	}
	defer fileOut.Close()
	// var CompanyMapOut CompanyMapOut

	companyMapOut := NewCompanyListOut(companyListIn)
	// for _, value := range companyMapOut.CompanySet {
	// 	fmt.Printf("%+v\n", value)
	// }

	dataToWrite, _ := json.Marshal(companyMapOut.CompanySet)
	fileOut.Write(dataToWrite)
	// for _, value := range companyMapOut.CompanySet {
	// 	dataToWrite, _ := json.Marshal(value)
	// 	fileOut.Write(dataToWrite)
	// }
}
