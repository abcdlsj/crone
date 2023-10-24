# Crone

`Go`'s cron expression parser library and cron scheduler.

## Usage


### Expr 
```go
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
```

### Scheduler
```go
package main

import (
	"fmt"
	"time"

	"github.com/abcdlsj/crone"
)

func main() {
	s := crone.NewSchduler()

	s.Add("job1", "*/1 * * * *", func() {
		fmt.Printf("job1: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	})

	s.Add("job2", "*/2 * * * *", func() {
		fmt.Printf("job2: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	})

	s.StartWithSignalListen()

	s.Wait()
}
```