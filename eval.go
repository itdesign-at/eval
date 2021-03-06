package eval

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var FloatError = math.NaN()

//
// Eval is the main struct converting an input string into an expression.
// It is a simple interpreter, that translates a calculation string into
// a float64, string or bool result.
//
// Example - used as plain golang code:
//  e := eval.New("(1+4) * (2-6) - 0.2")
//  _ = e.Parse()
//  r := e.Run() // r = -20.2
//
// Calculations:
//  +, -, *, /
//
type Eval struct {
	input     string
	exp       ast.Expr
	variables map[string]interface{}
}

// New is the main entry point with a calculation string to eval
//
// Example usage:
//  e := eval.New("round(10 * pow(2,2) + 3.141,2)")
//  if e.ParseExpr() == nil {
//    // prints "Result: 43.14"
//    fmt.Println("Result:", e.Run())
//  }
func New(input string) *Eval {
	var e Eval
	e.input = input
	return &e
}

// SetInput is used in unit tests to add another eval string
func (e *Eval) SetInput(input string) {
	e.input = input
}

// Variables adds external variables. In most cases these
// are float64 or strings.
func (e *Eval) Variables(variables map[string]interface{}) *Eval {
	e.variables = variables
	return e
}

// ParseExpr takes the input line and extracts tokens
func (e *Eval) ParseExpr() (err error) {
	e.exp, err = parser.ParseExpr(e.input)
	return
}

// Run returns the evaluated result or <nil> when nothing is wanted back
func (e *Eval) Run() interface{} {
	result := e.eval(e.exp)
	return result
}

// eval is the recursive interpreter
func (e *Eval) eval(exp ast.Expr) interface{} {
	switch exp := exp.(type) {
	// e.g. -17
	case *ast.UnaryExpr:
		switch exp.Op {
		case token.ADD:
			x := e.eval(exp.X)
			switch x.(type) {
			case int:
				return x.(int)
			case float64:
				return x.(float64)
			}
			return FloatError
		case token.SUB:
			x := e.eval(exp.X)
			switch x.(type) {
			case int:
				return -1 * x.(int)
			case float64:
				return -1 * x.(float64)
			}
			return FloatError
		}
	// ( expr )
	case *ast.ParenExpr:
		return e.eval(exp.X)
	// +, -, *, /
	case *ast.BinaryExpr:
		return e.evalBinaryExpr(exp)
	// token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
	case *ast.BasicLit:
		switch exp.Kind {
		case token.INT:
			i, _ := strconv.Atoi(exp.Value)
			return i
		case token.FLOAT:
			f, _ := strconv.ParseFloat(exp.Value, 64)
			return f
		case token.STRING:
			return exp.Value
		}
	// function calls
	case *ast.CallExpr:
		// alphabetically list of functions
		name := e.evalFunctionName(exp.Fun)
		switch name {
		case "abs":
			return e.abs(exp)
		case "avg":
			return e.avg(exp)
		case "env":
			return e.env(exp)
		case "float64":
			return e.float64(exp)
		case "ifExpr":
			return e.ifExpr(exp)
		case "int":
			return e.int(exp)
		case "isBetween":
			return e.isBetween(exp)
		case "isNaN":
			return e.isNaN(exp)
		case "max":
			return e.max(exp)
		case "min":
			return e.min(exp)
		case "pow":
			return e.pow(exp)
		case "regexpMatch":
			return e.regexpMatch(exp)
		case "round":
			return e.round(exp)
		case "setVal":
			return e.setVal(exp)
		case "sqrt":
			return e.sqrt(exp)
		case "substr":
			return e.substr(exp)
		case "sprintf":
			return e.sprintf(exp)
		case "time":
			return e.time(exp)
		case "val":
			return e.val(exp)
		default:
			return FloatError
		}
	case *ast.Ident:
		if exp.Name == "true" {
			return true
		}
		if exp.Name == "false" {
			return false
		}
		if val, ok := e.variables[exp.Name]; ok {
			return val
		}
	}

	return FloatError
}

// abs - implements the 'abs(x)' function and returns the absolute value of x.
// Returns a float64 value or math.NaN() on error.
func (e *Eval) abs(exp *ast.CallExpr) float64 {
	if len(exp.Args) != 1 {
		return FloatError
	}
	x := e.getArg(exp.Args[0])
	switch val := x.(type) {
	case int:
		return math.Abs(float64(val))
	case float64:
		return math.Abs(val)
	case string:
		val = stringer(val)
		float, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return math.Abs(float)
		}
	}
	return FloatError
}

// avg - implements the 'avg(x,y,z,...)' function and returns the average of a range numbers
// Returns a float64 value or math.NaN() on error.
func (e *Eval) avg(exp *ast.CallExpr) float64 {
	return e.avgMaxMin(exp, 3)
}

// env - implements the 'env("str")' function, reads the environment variable "str" and
// returns it's content as string.
func (e *Eval) env(exp *ast.CallExpr) string {
	l := len(exp.Args)
	if l < 1 {
		return ""
	}
	s := e.eval(exp.Args[0])
	var envResult string
	switch val := s.(type) {
	case string:
		val = stringer(val)
		envResult = os.Getenv(val)
	default:
	}
	return envResult
}

// float64 - implements the 'float64(x)' float64(x) function and converts x to float64
// Returns a float64 value or math.NaN() on error.
func (e *Eval) float64(exp *ast.CallExpr) float64 {
	l := len(exp.Args)
	if l < 1 {
		return FloatError
	}
	s := e.eval(exp.Args[0])
	// Attention! Check all basic numeric types - they could be in variables!
	switch val := s.(type) {
	case bool:
		if s.(bool) {
			return 1.0
		}
		return 0.0
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		val = stringer(val)
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f
		}
	default:
	}
	return FloatError
}

// ifExpr - implements 'if (<condition>,<true value>,<false value>)' which is
// similar to an 'if' statement in a programming language.
// Returns true/false or a math.NaN() on error.
func (e *Eval) ifExpr(exp *ast.CallExpr) interface{} {
	if len(exp.Args) != 3 {
		return FloatError
	}
	condition := e.getArg(exp.Args[0])
	trueValue := e.getArg(exp.Args[1])
	falseValue := e.getArg(exp.Args[2])
	switch condition.(type) {
	case bool:
		if condition.(bool) {
			if strVal, ok := trueValue.(string); ok {
				return stringer(strVal)
			}
			return trueValue
		}
		if strVal, ok := falseValue.(string); ok {
			return stringer(strVal)
		}
		return falseValue
	default:
	}
	return FloatError
}

// isBetween - implements 'isBetween(<val>,from,to)' where <val> must be string or float64
//
// Example:
//   isBetween(env("F"),49.0,51.0) ... checks if environment variable F >= 49.0 && F <= 51.0
//
// Returns true/false or a math.NaN() on error.
func (e *Eval) isBetween(exp *ast.CallExpr) interface{} {

	if len(exp.Args) != 3 {
		return false
	}

	// f64Value converts theValue to float64
	var f64Value = func(theValue interface{}) float64 {
		switch v := theValue.(type) {
		case int:
			return float64(v)
		case string:
			s := stringer(v)
			if s == "" {
				return FloatError
			}
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				if math.IsNaN(f) || math.IsInf(f, 0) {
					return FloatError
				}
				return f
			}
			return FloatError
		case float64:
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return FloatError
			}
			return v
		default:
			return FloatError
		}
	}

	var f64, from, to float64

	theValue := e.getArg(exp.Args[0])
	fromValue := e.getArg(exp.Args[1])
	toValue := e.getArg(exp.Args[2])

	f64 = f64Value(theValue)
	from = f64Value(fromValue)
	to = f64Value(toValue)

	return f64 >= from && f64 <= to
}

// isNaN - implements 'isNaN(<val>)' where <val> could be a valid float.
// This function is usable for error handling.
// Returns true or false.
func (e *Eval) isNaN(exp *ast.CallExpr) bool {
	if len(exp.Args) != 1 {
		return true
	}

	s := e.eval(exp.Args[0])
	// Attention! Check all basic numeric types - they could be in variables!
	switch val := s.(type) {
	case bool:
		return false
	case int:
		return math.IsNaN(float64(val))
	case int8:
		return math.IsNaN(float64(val))
	case int16:
		return math.IsNaN(float64(val))
	case int32:
		return math.IsNaN(float64(val))
	case int64:
		return math.IsNaN(float64(val))
	case uint:
		return math.IsNaN(float64(val))
	case uint8:
		return math.IsNaN(float64(val))
	case uint16:
		return math.IsNaN(float64(val))
	case uint32:
		return math.IsNaN(float64(val))
	case uint64:
		return math.IsNaN(float64(val))
	case float32:
		return math.IsNaN(float64(val))
	case float64:
		return math.IsNaN(val)
	case string:
		val = stringer(val)
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return true
		}
		return math.IsNaN(f)
	default:
		//
	}
	return true
}

// max returns the maximum of a range of numbers
// Returns float64 or a math.NaN() on error.
func (e *Eval) max(exp *ast.CallExpr) float64 {
	return e.avgMaxMin(exp, 2)
}

// min returns the minimum of a range of numbers
// Returns float64 or a math.NaN() on error.
func (e *Eval) min(exp *ast.CallExpr) float64 {
	return e.avgMaxMin(exp, 1)
}

func (e *Eval) avgMaxMin(exp *ast.CallExpr, flag int) float64 {
	if len(exp.Args) == 0 {
		return FloatError
	}

	var floats []float64

	for _, x := range exp.Args {
		f := e.getArg(x)
		switch val := f.(type) {
		case int:
			floats = append(floats, float64(val))
		case float64:
			floats = append(floats, val)
		case string:
			val = stringer(val)
			f := toFloat(val)
			if !math.IsNaN(f) { // skip invalid strings
				floats = append(floats, f)
			}
		}
	}

	if len(floats) < 1 {
		return FloatError
	}

	var val float64

	switch flag {
	case 1:
		val = floats[0]
		for i := 1; i < len(floats); i++ {
			val = math.Min(val, floats[i])
		}
	case 2:
		val = floats[0]
		for i := 1; i < len(floats); i++ {
			val = math.Max(val, floats[i])
		}
	case 3:
		for _, f := range floats {
			val = val + f
		}
		val = val / float64(len(floats))
	}

	return val
}

// pow - implements 'pow(<base x>,<exponent y>)' and returns x**y, the base-x exponential of y.
// Returns a float64 value or a math.NaN() on error.
func (e *Eval) pow(exp *ast.CallExpr) float64 {
	if len(exp.Args) != 2 {
		return FloatError
	}

	a := e.getArg(exp.Args[0])
	b := e.getArg(exp.Args[1])

	var fa, fb float64

	switch v := a.(type) {
	case int:
		fa = float64(v)
	case float64:
		fa = v
	case string:
		v = stringer(v)
		fa = toFloat(v)
	default:
		fa = FloatError
	}
	switch v := b.(type) {
	case int:
		fb = float64(v)
	case float64:
		fb = v
	case string:
		v = stringer(v)
		fb = toFloat(v)
	default:
		fb = FloatError
	}

	return math.Pow(fa, fb)
}

// regexpMatch - implements 'regexpMatch ("<regex>","string")' and returns true when the
// string matches
func (e *Eval) regexpMatch(exp *ast.CallExpr) bool {
	if len(exp.Args) != 2 {
		return false
	}
	var tmp interface{}
	var regexPattern string
	var regexString string
	tmp = e.getArg(exp.Args[0])
	switch val := tmp.(type) {
	case string:
		regexPattern = val
	default:
		return false
	}

	tmp = e.getArg(exp.Args[1])
	switch val := tmp.(type) {
	case string:
		regexString = val
	case int:
		regexString = fmt.Sprintf("%d", val)
	case bool:
		if tmp.(bool) {
			regexString = "true"
		} else {
			regexString = "false"
		}
	case float64:
		regexString = strconv.FormatFloat(tmp.(float64), 'f', -1, 64)
	default:
		return false
	}

	r, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}
	b := r.MatchString(regexString)
	return b
}

// round - implements the 'round (x,y)' function which
// rounds x to y decimal places.
//
// Returns a float64 value or math.NaN() on error.
func (e *Eval) round(exp *ast.CallExpr) float64 {
	if len(exp.Args) != 2 {
		return FloatError
	}

	a := e.getArg(exp.Args[0])
	b := e.getArg(exp.Args[1])

	var fa, fb float64

	switch v := a.(type) {
	case int:
		fa = float64(v)
	case float64:
		fa = v
	case string:
		fa = toFloat(v)
	default:
		fa = FloatError
	}
	switch v := b.(type) {
	case int:
		fb = float64(v)
	case float64:
		fb = v
	case string:
		fb = toFloat(v)
	default:
		fb = FloatError
	}

	x := math.Pow10(int(fb))

	return math.Round(fa*x) / x
}

// setVal - implements the 'setVal(a,b,c,d,...)' function which
// sets variables in pairs of 2.
// Returns nil or a golang error.
func (e *Eval) setVal(exp *ast.CallExpr) error {
	l := len(exp.Args)
	for i := 0; i < l; i++ {
		x := e.getArg(exp.Args[i])
		if i+1 < l {
			var name string
			var ok bool
			// name holds the variable name
			if name, ok = x.(string); !ok {
				continue
			}
			if e.variables == nil {
				e.variables = make(map[string]interface{})
			}
			name = stringer(name)
			if name == "" {
				continue
			}
			// value holds the variable value
			value := e.getArg(exp.Args[i+1])
			i += 1
			switch v := value.(type) {
			case string:
				v = stringer(v)
				e.variables[name] = v
			case bool, int, float64:
				e.variables[name] = v
			}
		}
	}
	return nil
}

// sqrt - implements 'sqrt(x)' which returns the square root of x.
// Returns a float64 value or math.NaN() on error.
func (e *Eval) sqrt(exp *ast.CallExpr) float64 {
	if len(exp.Args) != 1 {
		return FloatError
	}
	x := e.getArg(exp.Args[0])
	switch f := x.(type) {
	case int:
		return math.Sqrt(float64(f))
	case float64:
		return math.Sqrt(f)
	case string:
		f = stringer(f)
		return math.Sqrt(toFloat(f))
	default:
		return FloatError
	}
}

// substr - implements 'substr (string,start,size)' to get a piece of a string
//
// Examples:
//   substr("MyNameIsJohn",0,2)   ... "My"
//   substr("MyNameIsJohn",2,-1)  ... returns "NameIsJohn"
//   substr("MyNameIsJohn",-2,-1) ... returns "hn"
//   substr("MyNameIsJohn",-4,1)  ... returns "J"
//
// Returns a string or an empty string on error.
func (e *Eval) substr(exp *ast.CallExpr) string {
	const StringError = ""
	if len(exp.Args) != 3 {
		return StringError
	}
	theString := e.getArg(exp.Args[0])
	startPos := e.getArg(exp.Args[1])
	cutLen := e.getArg(exp.Args[2])
	switch theString.(type) {
	case string:
		s := theString.(string)
		if s == "" {
			return ""
		}
		var startP int
		var cutL int
		switch sp := startPos.(type) {
		case int:
			startP = sp
		case float64:
			startP = int(sp)
		}
		switch cutLen.(type) {
		case int:
			cutL = cutLen.(int)
		case float64:
			cutL = int(cutLen.(float64))
		}
		if cutL == 0 {
			return ""
		}
		if cutL > len(s) {
			cutL = len(s)
		}
		if math.Abs(float64(startP)) >= float64(len(s)) {
			return StringError
		}
		if startP >= 0 && cutL == -1 {
			return s[startP:]
		}
		l := len(s)
		if startP < 0 {
			if cutL == -1 {
				// e.g. last3 := s[len(s)-3:]
				return s[l+startP:]
			}
			x := l + startP
			if x+cutL >= l {
				cutL = l - x
			}
			return s[x : x+cutL]
		}
		if startP+cutL < startP {
			return StringError
		}
		if startP+cutL >= l {
			cutL = l - startP
		}
		return s[startP : startP+cutL]
	default:
	}
	return StringError
}

// time - implements 'time ("<action>","<format>")' to get a time as int64 or string
// Returns an int64 value or a string.
func (e *Eval) time(exp *ast.CallExpr) interface{} {
	if len(exp.Args) != 2 {
		return ""
	}

	a := e.getArg(exp.Args[0])
	b := e.getArg(exp.Args[1])

	switch left := a.(type) {
	case string:
		switch stringer(left) {
		case "", "now":
			switch right := b.(type) {
			case string:
				switch stringer(right) {
				case "", "epoch":
					return time.Now().Unix()
				case "rfc3339", "RFC3339":
					return time.Now().Format(time.RFC3339)
				}
			}
		case "starttime":
			var t time.Time
			// global.X.Lock()
			// t = global.X.ProgramStartTime
			// global.X.Unlock()
			switch right := b.(type) {
			case string:
				switch stringer(right) {
				case "", "epoch":
					return t.Unix()
				case "rfc3339", "RFC3339":
					return t.Format(time.RFC3339)
				}
			}
		}
	}
	return ""
}

// val - implements 'val("<name>")' to get the content of a variable. It returns
// an empty string when the variable is not found. Stored internally in the
// e.Variables(map[string]interface{}) map.
//
// Returns the value of the variable or an empty string on error.
func (e *Eval) val(exp *ast.CallExpr) interface{} {
	if len(exp.Args) != 1 || e.variables == nil {
		return ""
	}
	s := e.eval(exp.Args[0])
	if name, ok := s.(string); ok {
		key := stringer(name)
		if f, ok := e.variables[key]; ok {
			return f
		}
	}
	return ""
}

func (e *Eval) getArg(exp ast.Expr) interface{} {
	x := e.eval(exp)
	switch val := x.(type) {
	case bool:
		return val
	case int:
		return val
	case float64:
		return val
	case string:
		return stringer(val)
	default:
	}
	return math.NaN()
}

func (e *Eval) evalFunctionName(exp ast.Expr) string {
	return exp.(*ast.Ident).Name
}

func (e *Eval) evalBinaryExpr(exp *ast.BinaryExpr) interface{} {

	left := e.getArg(exp.X)
	right := e.getArg(exp.Y)

	switch exp.Op {
	case token.ADD:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 + 2
				return l + r
			case float64: // 1 + 3.141
				return float64(l) + r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 + 1
				return l + float64(r)
			case float64: // 3.141 + 3.141
				return l + r
			}
		}
	case token.SUB:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 - 2
				return l - r
			case float64: // 1 - 3.141
				return float64(l) - r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 - 1
				return l - float64(r)
			case float64: // 3.141 - 3.141
				return l - r
			}
		}
	case token.MUL:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 * 2
				return l * r
			case float64: // 1 * 3.141
				return float64(l) * r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 * 1
				return l * float64(r)
			case float64: // 3.141 * 3.141
				return l * r
			}
		}
	case token.QUO:
		// Divisions Ergebnis wird automatisch auf float64 gecastet
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 / 2
				if r == 0 {
					return math.Inf(1)
				}
				return float64(l) / float64(r)
			case float64: // 1 / 3.141
				if r == 0 {
					return math.Inf(1)
				}
				return float64(l) / r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 / 1
				if r == 0 {
					return math.Inf(1)
				}
				return l / float64(r)
			case float64: // 3.141 / 3.141
				if r == 0 {
					return math.Inf(1)
				}
				return l / r
			}
		}
	case token.EQL:
		switch l := left.(type) {
		case bool:
			switch r := right.(type) {
			case bool: // true == true
				return l == r
			}
		case int:
			switch r := right.(type) {
			case int: // 1 / 2
				return l == r
			case float64: // 1 / 3.141
				return float64(l) == r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 / 1
				return l == float64(r)
			case float64: // 3.141 / 3.141
				return l == r
			}
		case string:
			switch r := right.(type) {
			case string: // "strA" == "strB"
				return l == r
			}
		}
	case token.LSS:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 < 2
				return l < r
			case float64: // 1 < 3.141
				return float64(l) < r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 < 1
				return l < float64(r)
			case float64: // 3.141 < 3.141
				return l < r
			}
		}
	case token.GTR:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 > 2
				return l > r
			case float64: // 1 > 3.141
				return float64(l) > r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 > 1
				return l > float64(r)
			case float64: // 3.141 > 3.141
				return l > r
			}
		}
	case token.NEQ:
		switch l := left.(type) {
		case bool:
			switch r := right.(type) {
			case bool: // true != false
				return l != r
			}
		case int:
			switch r := right.(type) {
			case int: // 1 != 2
				return l != r
			case float64: // 1 != 3.141
				return float64(l) != r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 != 1
				return l == float64(r)
			case float64: // 3.141 != 3.141
				return l != r
			}
		case string:
			switch r := right.(type) {
			case string: // "strA" != "strB"
				return l != r
			}
		}
	case token.LEQ:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 <= 2
				return l <= r
			case float64: // 1 <= 3.141
				return float64(l) <= r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 <= 1
				return l <= float64(r)
			case float64: // 3.141 <= 3.141
				return l <= r
			}
		}
	case token.GEQ:
		switch l := left.(type) {
		case int:
			switch r := right.(type) {
			case int: // 1 >= 2
				return l >= r
			case float64: // 1 >= 3.141
				return float64(l) >= r
			}
		case float64:
			switch r := right.(type) {
			case int: // 3.141 >= 1
				return l >= float64(r)
			case float64: // 3.141 >= 3.141
				return l >= r
			}
		}
	case token.LAND:
		switch l := left.(type) {
		case bool:
			switch r := right.(type) {
			case bool: // true && false
				return l && r
			}
			//case int:
			//	switch r := right.(type) {
			//	case int: // 1 && 2
			//		return l && r
			//	case float64: // 1 && 3.141
			//		return float64(l) && r
			//	}
			//case float64:
			//	switch r := right.(type) {
			//	case int: // 3.141 && 1
			//		return l == float64(r)
			//	case float64: // 3.141 && 3.141
			//		return l && r
			//	}
			//case string:
			//	switch r := right.(type) {
			//	case string: // "strA" && "strB"
			//		return l && r
			//	}
		}
	case token.LOR:
		switch l := left.(type) {
		case bool:
			switch r := right.(type) {
			case bool: // true || true
				return l || r
			}
			//case int:
			//	switch r := right.(type) {
			//	case int: // 1 || 2
			//		return l || r
			//	case float64: // 1 / 3.141
			//		return float64(l) || r
			//	}
			//case float64:
			//	switch r := right.(type) {
			//	case int: // 3.141 || 1
			//		return l || float64(r)
			//		//case float64: // 3.141 || 3.141
			//		//	return l || r
			//	case string:
			//		switch r := right.(type) {
			//		case string: // "strA" || "strB"
			//			return l || r
			//		}
			//	}
		}
	case token.OR:
		switch l := left.(type) {
		//case bool:
		//	switch r := right.(type) {
		//	case bool: // true | true
		//		return l | r
		//	}
		case int:
			switch r := right.(type) {
			case int: // 1 | 2
				return l | r
				//case float64: // 1 / 3.141
				//	return float64(l) | r
			}
			//case float64:
			//	switch r := right.(type) {
			//	case int: // 3.141 | 1
			//		return l | float64(r)
			//		case float64: // 3.141 | 3.141
			//			return l | r
			//	case string:
			//		switch r := right.(type) {
			//		case string: // "strA" | "strB"
			//			return l | r
			//		}
			//	}
		}
	case token.AND:
		switch l := left.(type) {
		//case bool:
		//	switch r := right.(type) {
		//	case bool: // true & true
		//		return l & r
		//	}
		case int:
			switch r := right.(type) {
			case int: // 1 & 2
				return l & r
				//case float64: // 1 & 3.141
				//	return float64(l) & r
			}
			//case float64:
			//	switch r := right.(type) {
			//	case int: // 3.141 & 1
			//		return l & float64(r)
			//	case float64: // 3.141 & 3.141
			//		return l & r
			//	case string:
			//		switch r := right.(type) {
			//		case string: // "strA" & "strB"
			//			return l & r
			//		}
			//	}
		}
	}
	return FloatError
}

func (e *Eval) sprintf(exp *ast.CallExpr) interface{} {
	l := len(exp.Args)
	switch l {
	case 0:
		return FloatError
	case 1:
		if format, ok := e.getArg(exp.Args[0]).(string); ok {
			return format
		}
	default:
		var format = ""
		var params []interface{}
		format, _ = e.getArg(exp.Args[0]).(string)
		for i := 1; i < l; i++ {
			params = append(params, e.eval(exp.Args[i]))
		}
		return fmt.Sprintf(format, params...)
	}
	return FloatError
}

// int converts input to an integer
func (e *Eval) int(exp *ast.CallExpr) interface{} {
	l := len(exp.Args)
	if l < 1 {
		return FloatError
	}
	s := e.eval(exp.Args[0])
	// Attention! Check all basic numeric types - they could be in variables!
	switch val := s.(type) {
	case bool:
		if s.(bool) {
			return 1
		}
		return 0
	case int:
		return val
	case int8:
		return int(val)
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case uint:
		return int(val)
	case uint8:
		return int(val)
	case uint16:
		return int(val)
	case uint32:
		return int(val)
	case uint64:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case string:
		val = stringer(val)
		i, err := strconv.Atoi(val) // first try -> integer
		if err == nil {
			return i
		}
		f, err := strconv.ParseFloat(val, 64) // second try -> float64
		if err == nil {
			return int(f)
		}
	default:
	}
	return FloatError
}

// stringer removes "" from a string at the beginning and at the end
func stringer(s string) string {
	if len(s) < 1 {
		return ""
	}
	if s[0:1] == `"` && s[len(s)-1:] == `"` {
		return strings.Trim(s, `"`)
	}
	return s
}

// toFloat takes string s and converts it to a float64 value. It
// returns FloatError on error which can be checked with math.IsNaN(f).
func toFloat(s string) float64 {
	var err error
	var i int
	var f float64
	i, err = strconv.Atoi(s)
	if err == nil {
		return float64(i)
	}
	f, err = strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	}
	return FloatError
}
