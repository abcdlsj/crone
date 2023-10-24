package main

import (
	"fmt"
	"time"

	"github.com/abcdlsj/crone"
)

func main() {
	rule := "* * * * *"
	cron := crone.NewExpr(rule)
	fmt.Println(cron.Next(time.Now()))
}
