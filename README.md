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
	
	// .Variables adds the golang map from above as variables
	e := eval.New(input).Variables(variables)

	// ParseExpr() must be OK to continue
	if e.ParseExpr() == nil {
		output := e.Run()
		// prints "(true) the result is: 98.66"
		fmt.Println(output)
	}
}
```
# Variables
# Functions
## abs(x) 
abs implements the 'abs(x)' function and returns the absolute value of x.

    abs(-3.14)   ... 3.14 // float64 as input
    abs(3.14)    ... 3.14
    abs("-2.55") ... 2.55 // strings ok when numeric
    abs(-2)      ... 2    // int

Returns a float64 value or math.NaN() on error.

## avg(x,y,z,...)
avg implements the 'avg(x,y,z,...)' function and returns the average of a range of numbers

    avg(10,20) ... 15.0 // numbers only
    avg(30,"10","20.0","John Doe")` ... 20.0 // mixes input "John Doe" is ignored

Returns a float64 value or math.NaN() on error.

## env("str")
env - implements the 'env("str")' function, reads the environment variable "str" and
returns it's content as string.
The main purpose of reading environment variables is to make it possible to pass something
from the outside when calling the main program. 

    env("HOME") ... e.g. root under linux
    float64(env("pi")) ... 3.14159 as float64 when 'pi' is set
    ifExpr(env("notSet")=="","isEmpty","isFilled") ... "isEmpty" as string

Returns an empty string when not found.