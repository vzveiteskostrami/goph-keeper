package dialog

import (
	"fmt"
	"os"
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
