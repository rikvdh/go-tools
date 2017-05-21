package main

import (
	"flag"
	"fmt"

	"github.com/rikvdh/go-tools/lib/brightness"
)

func main() {
	newBrightness := flag.Float64("s", -1, "Brightness in %")
	flag.Parse()

	b, err := brightness.New()
	if err != nil {
		panic(err)
	}


	if *newBrightness < 0 {
		fmt.Printf("Brightness is: %.1f\n", b.Get())
	} else {
		b.Set(*newBrightness)
	}
}
