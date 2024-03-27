package dialog

import (
	"fmt"
	"os"
	"strconv"
)

func GetAnswer(prompt string) string {
	var answer string
	for {
		fmt.Print(prompt)
		fmt.Fscan(os.Stdin, &answer)
		if answer != "" {
			break
		}
	}
	return answer
}

func Yn(prompt string) bool {
	var answer string
	for {
		fmt.Print(prompt + " (y/n)?")
		fmt.Fscan(os.Stdin, &answer)
		if answer == "y" || answer == "n" {
			break
		}
		answer = ""
	}
	return answer == "y"
}

func GetInt(prompt string, usemin bool, min int64, usemax bool, max int64) (int64, bool) {
	var answer string
	var r int64
	var err error
	for {
		fmt.Print(prompt)
		fmt.Fscan(os.Stdin, &answer)
		if answer != "" {
			if answer == "-" {
				return 0, false
			}
			r, err = strconv.ParseInt(answer, 10, 64)
			if err != nil {
				fmt.Printf("Не удалось преобразовать %s в число.\n\r", answer)
			} else {
				b := true
				if usemin {
					if r < min {
						fmt.Printf("Значение должно быть больше либо равно %d.\n\r", min)
						b = false
					}
				}
				if usemin {
					if r > max {
						fmt.Printf("Значение должно быть меньше либо равно %d.\n\r", max)
						b = false
					}
				}
				if b {
					break
				}
			}
		}
	}
	return r, true
}
