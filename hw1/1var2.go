package main

import "fmt"

type Mod func(sandglass [][]string) [][]string

const (
	SIZE = 15
	RED  = "\x1B[31m"
	CHAR = "X"
)

func createDefault() [][]string {
	sandglass := make([][]string, SIZE)
	for i := range sandglass {
		sandglass[i] = make([]string, SIZE)
	}
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if i == 0 || i == SIZE-1 || i == j || i == SIZE-j-1 {
				sandglass[i][j] = CHAR
			} else {
				sandglass[i][j] = " "
			}
		}
	}
	return sandglass
}

func myPrint(array [][]string) {
	for i := range array {
		for _, value := range array[i] {
			fmt.Print(value)
		}
		fmt.Println()
	}
}

func changeChar(newChar string) Mod {
	return func(sandglass [][]string) [][]string {
		for i := range sandglass {
			for j := range sandglass[i] {
				if i == 0 || i == len(sandglass[i])-1 || i == j || i == len(sandglass[i])-j-1 {
					sandglass[i][j] = newChar
				}
			}
		}
		return sandglass
	}
}

func changeColor(newColor string) Mod {
	return func(sandglass [][]string) [][]string {
		for i := range sandglass {
			for j := range sandglass[i] {
				if i == 0 || i == len(sandglass[i])-1 || i == j || i == len(sandglass[i])-j-1 {
					sandglass[i][j] = newColor + sandglass[i][j]
				}
			}
		}
		return sandglass
	}
}

func changeSize(newSize int) Mod {
	return func(sandglass [][]string) [][]string {
		currentChar := sandglass[0][0]
		newSandglass := make([][]string, newSize)
		for i := range newSandglass {
			newSandglass[i] = make([]string, newSize)
		}
		for i := range newSandglass {
			for j := range newSandglass[i] {
				if i == 0 || i == len(newSandglass)-1 || i == j || i == len(newSandglass)-j-1 {
					newSandglass[i][j] = currentChar
				} else {
					newSandglass[i][j] = " "
				}
			}
		}
		return newSandglass
	}
}

func sandglass(mods ...Mod) {
	sandglass := createDefault()
	for _, mod := range mods {
		sandglass = mod(sandglass)
	}
	myPrint(sandglass)
}

func main() {
	sandglass(changeSize(10), changeChar("!"), changeColor(RED))
}
