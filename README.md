# eval
go eval implementation to use and learn the ast package  

First example:
```
package main

import (
	"fmt"
	"github.com/itdesign-at/eval"
)

func main() {
	
	// add variables which can be used
	var variables = map[string]interface{}{
		"pi":        3.141,
		"text":      "the result is",
		"n":         10,
		"boolValue": true,
	}

	// what we want to eval
	const input = `sprintf ("(%v) %s: %.4g",boolValue,text,round(pow(pi,2)*n,2))`
	
	e := eval.New(input).Variables(variables)

	// ParseExpr() must be OK to continue
	if e.ParseExpr() == nil {
		output := e.Run()
		// prints "(true) the result is: 98.66"
		fmt.Println(output)
	}
}

```

