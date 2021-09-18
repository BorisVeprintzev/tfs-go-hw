//package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
)

type MyMap map[string]string
type ModString func(sandglass [][]string)

const (
	RED                = "\x1B[31m"
	SIZE        string = "size"
	CH          string = "ch"
	COLORSTRING        = "color"
	VAR1               = 1
	VAR2               = 2
)

const (
	NUMBER = iota
	CHARACTER
	COLOR
)

func addColor(color string, text string) {
	fmt.Printf("%s%s", color, text)
}

func print(char string, size int, color string) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i == 0 || i == size-1 || i == j || i == size-j-1 {
				addColor(char, color)
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Println()
	}
}

func sandglass() {

}

func sandglass1(myMap MyMap) {
	size, err := strconv.Atoi(myMap[SIZE])
	if err != nil {
		log.Fatal(err)
	}
	print(myMap[CH], size, myMap[COLORSTRING])
}

// первая строка кол-во, 2 символы, 3 цвет
func sandglass2(values []string) {
	number, err := strconv.Atoi(values[NUMBER])
	if err != nil {
		log.Fatal(err)
	}
	print(values[CHARACTER], number, values[COLOR])
}

func main() {
	size := flag.Int("n", 15, "size of sandglass")
	ch := flag.String("ch", "X", "symbol of sandglass")
	color := flag.String("color", RED, "color of text")
	funcNumber := flag.Int("var", 1, "func var 1 or 2")
	flag.Parse()

	switch *funcNumber {
	case VAR1:
		myMap := make(MyMap)
		myMap[SIZE] = strconv.Itoa(*size)
		myMap[CH] = *ch
		myMap[COLORSTRING] = *color
		sandglass1(myMap)
	case VAR2:
		var data [3]string
		data[0] = strconv.Itoa(*size)
		data[1] = *ch
		data[2] = *color
		sandglass2(data[:])
	default:
		fmt.Println("Введен неверный вариант. Выход.")
	}
}
