package emp

import (
	"fmt"
)

func BlockingLogger(channel chan string) {
	var log string
	for {
		log = <-channel
		fmt.Println(log)
		if log == "Quit" {
			break
		}
	}
}
