package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	EnvName = "ENV FILE"
)

type InInfoCompany struct {
	Company       string      `json:"company"`
	OperationType string      `json:"type"`
	Value         interface{} `json:"value"`
	Id            interface{} `json:"id"`
	CreatedAt     interface{} `json:"created_at"`
	Operation
}

type Operation struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	Id        interface{} `json:"id"`
	CreatedAt interface{} `json:"created_at"`
}

type OutInfoCompany struct {
	Company             string   `json:"company"`
	ValidCountOperation int      `json:"valid_operations_count"`
	Balance             int      `json:"balance"`
	InvalidOperation    []string `json:"invalid_operations"`
}

func NewOutInfoCompany(name string, validCountOp int, balance int, invalidOp []string) OutInfoCompany {
	return OutInfoCompany{
		Company:             name,
		ValidCountOperation: validCountOp,
		Balance:             balance,
		InvalidOperation:    invalidOp,
	}
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
	fmt.Scanf("%s", *name)
	return *name
}

func main() {
	filename := readFileName()
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Can't to open file:", filename, err)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	var companyList []InInfoCompany
	for {
		var company InInfoCompany
		if err := decoder.Decode(&company); err != io.EOF {
			break
		} else if err != nil {
			log.Fatal("Can't decode.", err)
		}
		fmt.Println(company.Company)
		companyList = append(companyList, company)
	}
	fmt.Println(companyList)
}
