# eval
go eval implementation to use and learn the ast package. A working shell calculator which uses
this packages can be found in cmd/calc.

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
As in golang variables are written as character-strings but with the exception that special characters can be used, too.
See function val("var") and setVal("var") for details.

# Numeric calculations
Basic numeric calculations are implemented +, -, / and *. 

# Functions
Alphabetically list of function
## abs (x) 
abs implements the 'abs(x)' function and returns the absolute value of x.

    abs(-3.14)   ... 3.14 // float64 as input
    abs(3.14)    ... 3.14
    abs("-2.55") ... 2.55 // strings ok when numeric
    abs(-2)      ... 2    // int

Returns a float64 value or math.NaN() on error.

## avg (x,y,z,...)
avg implements the 'avg(x,y,z,...)' function and returns the average of a range of numbers

    avg(10,20) ... 15.0 // numbers only
    avg(30,"10","20.0","John Doe")` ... 20.0 // mixes input "John Doe" is ignored

Returns a float64 value or math.NaN() on error.

## env ("str")
env - implements the 'env("str")' function, reads the environment variable "str" and
returns it's content as string.
The main purpose of reading environment variables is to make it possible to pass something
from the outside when calling the main program. 

    env("HOME") ... e.g. root under linux
    float64(env("pi")) ... 3.14159 as float64 when 'pi' is set
    ifExpr(env("notSet")=="","isEmpty","isFilled") ... "isEmpty" as string

Returns an empty string when not found.

## float64 (x)
float64 - implements the 'float64(x)' function and converts x to float64

    float64(pow(2,2)) ... 4.0
    float64("-2.27")" ... -2.27  // string ok when numeric
    float64(1>0 && 2>1) ...  1.0 // bool true results in 1.0

Returns a float64 value or math.NaN() on error.

## ifExpr (condition,x,y)
ifExpr - implements 'if (condition,true value,false value)' which is
similar to an 'if' statement in a programming language. Can also be compared with
spreadsheets '=IF()' statement.

    ifExpr(x>1,100,0)                 ... depends on x, returns 100 or 0
    ifExpr(2>1,"greater 1","lower 1") ... returns "greater 1" as string
    ifExpr(2>1,1==1,1==0)             ... returns true as bool

Returns true/false or math.NaN() on error.

## int (x)
int - implements the 'int(x)' function and converts x to int

    int(-3.141) ... -3
    int("-1")`  ... -1
    int(false)` ... 0

## isBetween (x,a,z)
isBetween returns true if x >= a and x <= z, otherwise false

    isBetween(-1,0,1) ... false
    isBetween(-0.95,-0.99,-0.90) ... true
    isBetween(something,"Wrong",/) ... false

## isNaN (f)
isNaN - implements 'isNaN(f)' and checks if given f is a float64.

    isNaN(float64(NaN)) ... true
    isNaN(5.1) ... false

This function is usable for error handling.

## max (n1,n2,...)         
max returns the maximum of a range of numbers

    max(0,-3.33,97.77) ... 97.77
    max("10",8) ... 10

Returns the maximum as float64 value or math.NaN() on error.

## min (n1,n2,...)
min returns the minimum of a range of numbers

    min(0,-3.33,97.77) ... -3.33
    min("10",8) ... 8

Returns the minimum as float64 value or math.NaN() on error.

## pow (x,y)
pow returns x**y, the base-x exponential of y

    pow(2,0) ... 1
    pow(2,2) ... 4

Returns a float64 value or a math.NaN() on error.

## regexpMatch ("r","s")
regexpMatch checks string s against regular expression r

    regexpMatch ("^\d+$","1234") ... true

Returns true or false.

## round (x,y)
round x to y digits

    round(3.14159,3) ... 3.142
    round(3.14159,0) ... 3
    round(3.14159,-1) ... 0

Returns a float64 value or math.NaN() on error.

## setVal (pairs)
e.g. setVal("i",1,"s","str", etc.) set a range of variables (key -> value pairs)

    setVal("a",10,"$SYS/b",20) ... set a to 10 and $SYS/b to 20
    setVal("a",true) ... set boolean a to true

setVal allows to add variables with special characters in the key (see $SYS/b example)

## sqrt (x)
qrt - implements 'sqrt(x)' which returns the square root of x.

    sqrt(16) ... 4
    round(sqrt(3),2) ... 1.73

Returns a float64 value or a math.NaN() on error.

## substr("str",idx,len)
extract a substring out of "str"

    substr("MyNameIsJohn",0,1) ... M
    substr("MyNameIsJohn",2,4) ... Name
    substr("MyNameIsJohn",-4,-1) ... John

## time ("action","format")
time - implements 'time ("<action>","<format>")' to get a time as int64 or string

    time("","")                 ...  1423542512 = ("now","epoch") (int64)
    time("now","RFC3339")       ...  2020-07-02T07:39:10+02:00 (string)
    time("starttime","epoch")   ...  1423542512 (int64), start time of program
    time("starttime","rfc3339") ...  2020-07-02T07:39:10+02:00 (string)

Returns an int64 value or a string.

## val ("key")
val - implements 'val("key")' to get the content of a variable. It returns
an empty string when the variable is not found. 

    val("$SYS/b") ... value of variable $SYS/b when set (see funtion setVal())

Returns the value of the variable or an empty string on error.