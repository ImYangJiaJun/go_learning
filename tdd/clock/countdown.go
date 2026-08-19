package clock

import (
	"fmt"
	"io"
	"time"
)

func Countdown(out io.Writer, sleep func(time.Duration)) {
	for i := 3; i > 0; i-- {
		_, err := fmt.Fprintf(out, "%d\n", i)
		if err != nil {
			return
		}
		sleep(time.Second)
	}
	_, err := fmt.Fprint(out, "Go!\n")
	if err != nil {
		return
	}
}
