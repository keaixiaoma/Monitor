package log

import (
	"fmt"
	"time"
)

func ErrPrint(err error) {
	fmt.Printf("[Monitor]-[%s] Error: %s\n", time.Now(), err)
}

func Print(log string) {
	fmt.Printf("[Monitor]-[%s]  %s\n", time.Now(), log)
}
