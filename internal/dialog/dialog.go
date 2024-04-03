package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func GetAnswer(prompt string, secret bool, canBeEmpty bool) string {
	var answer string
	for {
		fmt.Print(prompt)
		if secret {
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return "-"
			}
			answer = string(bytePassword)
			fmt.Println("")
		} else {
			answer = getInput()
		}
		if canBeEmpty || answer != "" {
			break
		}
	}
	return answer
}

func Yn(prompt string) bool {
	var answer string
	for {
		fmt.Print(prompt + " (y/n)?")
		answer = getInput()
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
		if answer = getInput(); answer != "" {
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

func Menu(data []string) int16 {
	ln := len(data) - 1

	if ln < 1 {
		return 0
	}

	rg := 0
	for ln > 0 {
		ln = ln / 10
		rg++
	}

	delim := DrawHeader(data[0], true)
	for ln = 1; ln < len(data); ln++ {
		fmt.Print(strings.Repeat(" ", rg-len(strconv.Itoa(ln))))
		fmt.Printf("%d. ", ln)
		fmt.Println(data[ln])
	}
	fmt.Print(strings.Repeat(" ", rg-1))
	fmt.Println("-. Отказаться от продолжения")
	fmt.Println(delim)
	ans, _ := GetInt("Выберите:", true, 1, true, int64(len(data)-1))
	return int16(ans)
}

func DrawHeader(he string, full bool) string {
	delim := strings.Repeat("-", len([]rune(he))+3)
	fmt.Print(delim + "\n\r")
	fmt.Print("-- " + he + " \n\r")
	if full {
		fmt.Print(delim + "\n\r")
	}
	return delim
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimRight(answer, " \n\r")
	return answer
}
