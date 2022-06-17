package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/itdesign-at/eval"
)

/**

Shell calculator example

Example usage:

go build
./calc -n 16 -text "Shell calculator result:" -pi 3.141 'sprintf ("%s %.3f",text,pi*n)'
Shell calculator result: 50.256

*/

func main() {
	// last element of command line
	toEval := os.Args[len(os.Args)-1]

	// treat each argument as variable and add it
	opts := parse(os.Args)
	e := eval.New(toEval).Variables(opts)

	// execute it
	if err := e.ParseExpr(); err == nil {
		fmt.Println(e.Run())
	} else {
		log.Println(err.Error())
		os.Exit(1)
	}
}

// parse takes shell args and maps it to key/values
func parse(args []string) map[string]interface{} {
	var opt = make(map[string]interface{})
	n := len(args)
	if n < 2 {
		return opt
	}
	var key, value string
	for i := 1; i < n; i++ {
		key = ""
		if strings.HasPrefix(args[i], "-") {
			key = strings.TrimSpace(strings.TrimLeft(args[i], "-"))
		}
		if key == "" {
			continue
		}
		if i+1 == n { // end reached?
			opt[key] = true
			break
		}
		value = args[i+1]
		// first character is a mask character
		// e.g. -negative "\-3"
		if strings.HasPrefix(value, `\`) {
			value = value[1:]
			if value == "" {
				opt[key] = `\`
			} else {
				if f, err := strconv.ParseFloat(value, 64); err == nil {
					opt[key] = f
				} else {
					opt[key] = value
				}
			}
			i++
			continue
		}
		// if "-key1" follows "-key2"
		if strings.HasPrefix(value, `-`) {
			opt[key] = true
			continue
		}
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			opt[key] = f
		} else {
			switch value {
			case "true":
				opt[key] = true
			case "false":
				opt[key] = false
			default:
				opt[key] = value
			}
		}
		i++
	}
	return opt
}
