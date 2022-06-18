package eval

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

// TestEvalParser tests wrong input
func TestEvalParser(t *testing.T) {
	var wrong = []string{"", ";", ",", "'"}

	for _, w := range wrong {
		e := New(w)
		if e.ParseExpr() == nil {
			t.Errorf("Input %s should lead to an error", w)
		}
	}
}

// TestBit tests bit OR (|) and AND (&) operator with floating point values
func TestBit(t *testing.T) {

	var falseInput = map[string]interface{}{
		"3.0 | 3":    math.NaN(),
		"3 | 3.0":    math.NaN(),
		"3.0 & 3":    math.NaN(),
		"3 & 3.0":    math.NaN(),
		"true | 3":   math.NaN(),
		"3 | false":  math.NaN(),
		"\"x\" | 3":  math.NaN(),
		" 3 | \"x\"": math.NaN(),
	}

	for k := range falseInput {
		e := New(k)
		if e.ParseExpr() != nil {
			t.Errorf("ParseExpr %s leads to error %s", k, e)
		}
		r := e.Run()
		var f float64
		var ok bool
		if f, ok = r.(float64); !ok || !math.IsNaN(f) {
			t.Errorf("Input %s show result in NaN %f", k, f)
		}
	}

	var goodInput = map[string]int{
		"1 | 1": 1,
		"0 | 1": 1,
		"4 | 3": 7,
		"1 & 1": 1,
		"0 & 1": 0,
		"4 & 3": 0,
		"7 & 3": 3,
	}

	for k, v := range goodInput {
		e := New(k)
		if e.ParseExpr() != nil {
			t.Errorf("Input %s leads to an error", k)
		}
		r := e.Run()
		var f int
		var ok bool
		if f, ok = r.(int); !ok {
			t.Errorf("Input %s leads to an error", k)
		}
		if f != v {
			t.Errorf("Input %s leads to an error", k)
		}
	}
}

func TestVars(t *testing.T) {
	var input = map[string]interface{}{
		"host":     "www.orf.at",
		"n":        4,
		"Pi":       3.141,
		"$Sys/tmp": "Na bitte",
	}
	e := New(`sprintf ("%s %s %.3f",val("$Sys/tmp"),host,pow(n,Pi))`).Variables(input)
	err := e.ParseExpr()
	if err == nil {
		out := e.Run()
		if out != "Na bitte www.orf.at 77.816" {
			t.Errorf("got %s", out)
		}
	}
}

// TestSingleNumber coverts single strings to float64 values
func TestSingleNumber(t *testing.T) {

	var intInput = map[string]int{
		"1": 1,
		"0": 0,
	}

	for k, v := range intInput {
		e := New(k)
		if e.ParseExpr() != nil {
			t.Errorf("ParseExpr error for %s", k)
		}
		r := e.Run()
		var i int
		var ok bool
		if i, ok = r.(int); !ok {
			t.Errorf("Returned value for %s is not an integer", k)
		}
		if i != v {
			t.Errorf("Values for %s are different", k)
		}
	}

	var floatInput = map[string]float64{
		"1.0": 1.0,
		"0.0": 0.0,
	}

	for k, v := range floatInput {
		e := New(k)
		if e.ParseExpr() != nil {
			t.Errorf("ParseExpr error for %s", k)
		}
		r := e.Run()
		var f float64
		var ok bool
		if f, ok = r.(float64); !ok {
			t.Errorf("Returned value for %s is not a float", k)
		}
		if f != v {
			t.Errorf("Values for %s are different", k)
		}
	}

}

func TestDivZero(t *testing.T) {
	var divZero = map[string]float64{
		"0/0":          math.Inf(1),
		"0 / 0":        math.Inf(1),
		"-1 / 0":       math.Inf(1),
		"1.1 / 0":      math.Inf(1),
		"-1.1 / 0":     math.Inf(1),
		"-1.1 / (1-1)": math.Inf(1),
	}
	for s, r := range divZero {
		e := New(s)
		if e.ParseExpr() != nil {
			t.Errorf("Input %s leads to an error", s)
		}
		result := e.Run()
		if result != r {
			t.Errorf("Input %s leads to an error, result = %v but we expect %f", s, e.Run(), r)
		}
	}
}

func TestCalcsWithFloatInt(t *testing.T) {
	// WN: Bei der Division wird automatisch auf float64 gecastet
	var ok = map[string]float64{
		"1 / 1":                   1,
		"-1 / -1":                 1,
		"1 + 3.141":               4.141,
		"3.141 + 1":               4.141,
		"-18 / -10":               1.8,
		"round(-10/-18,\"2\")":    0.56,
		"round(-10/-18,2)":        0.56,
		"round(pow(-10/-18,2),2)": 0.31,
		"round(pow(2,-10/-18),2)": 1.47,
	}
	for s, r := range ok {
		e := New(s)
		if e.ParseExpr() != nil {
			t.Errorf("Input %s leads to an error", s)
		}
		result := e.Run()
		if result != r {
			t.Errorf("Input %s leads to an error, result = %v but we expect %f", s, e.Run(), r)
		}
	}
}

func TestCalcs(t *testing.T) {
	var intOks = map[string]int{
		"0 + 0":     0,
		"1 + 1":     2,
		"-1 + -1":   -2,
		"-18 + -10": -28,
		"0 - 0":     0,
		"1 - 1":     0,
		"-1 - -1":   0,
		"-18 - -10": -8,
		"0 * 0":     0,
		"1 * 1":     1,
		"-1 * -1":   1,
		"-18 * -10": 180,
	}

	for s, r := range intOks {
		e := New(s)
		if e.ParseExpr() != nil {
			t.Errorf("Input %s leads to an error", s)
		}
		result := e.Run()
		if result != r {
			t.Errorf("Input %s leads to an error, result = %v but we expect %d", s, e.Run(), r)
		}
	}
}

func TestCompare(t *testing.T) {

	var ok = map[string]bool{
		"3.141 == val(\"pi\")":    true,
		"val(\"pi\") == 3.141":    true,
		"val(\"i\") == 2":         true,
		"2 == 2":                  true,
		"1 == 1":                  true,
		"0 == 1":                  false,
		"\"a\" == \"a\"":          true,
		"\"a\" == \"b\"":          false,
		"1 != 1":                  false,
		"0 != 1":                  true,
		"\"a\" != \"a\"":          false,
		"\"a\" != \"b\"":          true,
		"1 == 1 && 2 == 2":        true,
		"1 == 1 && 2 == 3":        false,
		"1 == 1 || 2 == 3":        true,
		"1 == 2 || 2 == 3":        false,
		"1 == 2 || 2 != 3":        true,
		"\"a\" == \"a\" && 2 > 0": true,
		"\"a\" != \"a\" && 2 > 0": false,
		`(val("Rtt")>=0.03)`:      true,
	}

	for s, r := range ok {
		e := New(s)
		e.Variables(map[string]interface{}{"i": 2, "pi": 3.141, "Rtt": 0.046})
		if e.ParseExpr() != nil {
			t.Errorf("Input %s leads to an error", s)
		}
		if e.Run() != r {
			t.Errorf("Input %s leads to an error, result = %v but we expect %v", s, e.Run(), r)
		}
	}
}

func TestAbs(t *testing.T) {

	var ok = map[string]float64{
		"abs(\"3.14\")":          3.14,
		"abs(3.14)":              3.14,
		"abs(-3)":                3,
		"abs(-3.14)":             3.14,
		"abs(3.14) / abs(-3.14)": 1,
	}
	for s, r := range ok {
		e := New(s)
		if e.ParseExpr() != nil {
			t.Errorf("Input %s leads to an error", s)
		}
		if e.Run() != r {
			t.Errorf("Input %s leads to an error, result = %v but we expect %.2f", s, e.Run(), r)
		}
	}
}

func TestIfExpr(t *testing.T) {
	e := New(`ifExpr(val("n")>3||val("n")==2,66,55)`)
	err := e.ParseExpr()
	if err != nil {
		panic(err)
	}
	e.Variables(map[string]interface{}{"n": 2.0})
	result := e.Run()
	if result != 66 {
		t.Errorf("Expected 66 as output but got %v", result)
	}

	type X struct {
		str        string
		shouldBe   bool
		trueValue  interface{}
		falseValue interface{}
	}

	var testSuite = []X{
		{
			str:        `ifExpr(true,true,false)`,
			shouldBe:   true,
			trueValue:  true,
			falseValue: false,
		},
		{
			str:        `ifExpr(false,true,false)`,
			shouldBe:   false,
			trueValue:  true,
			falseValue: false,
		},
		{
			str:        `ifExpr(true,7*3,0)`,
			shouldBe:   true,
			trueValue:  21,
			falseValue: 0,
		},
		{
			str:        `ifExpr(false,7*3,-8*4`,
			shouldBe:   false,
			trueValue:  21,
			falseValue: -32,
		},
		{
			str:        `ifExpr(1 == 1,2-1,0)`,
			shouldBe:   true,
			trueValue:  1,
			falseValue: 0,
		},
		{
			str:        `ifExpr("a" == "a",pow(2,2),0)`,
			shouldBe:   true,
			trueValue:  4.0,
			falseValue: 0.0,
		},
		{
			str:        `ifExpr("a" != "a","OK",abs(-3.2))`,
			shouldBe:   false,
			trueValue:  "OK",
			falseValue: 3.2,
		},
		{
			str:        `ifExpr(1>0,1,0)`,
			shouldBe:   true,
			trueValue:  1,
			falseValue: 0,
		},
		{
			str:        `ifExpr(0>1,"OK","NOTOK")`,
			shouldBe:   false,
			trueValue:  "OK",
			falseValue: "NOTOK",
		},
		{
			str:        `ifExpr(1>=0,1,0)`,
			shouldBe:   true,
			trueValue:  1,
			falseValue: 0,
		},
		{
			str:        `ifExpr(1>=2,1,(-4+2)/sqrt(4))`,
			shouldBe:   false,
			trueValue:  1.0,
			falseValue: -1.0,
		},
		{
			str:        `ifExpr(1>=0,34/2,0)`,
			shouldBe:   true,
			trueValue:  17.0,
			falseValue: 0,
		},
		{
			str:        `ifExpr(1>=2,1,-4 * -2)`,
			shouldBe:   false,
			trueValue:  1,
			falseValue: 8,
		},
		{
			str:        `ifExpr(1>=2 || 2>-1,sqrt(64),-6)`,
			shouldBe:   true,
			trueValue:  8.0,
			falseValue: -6,
		},
		{
			str:        `ifExpr(1>=2 && pow(2,2)>-1,sqrt(64),-6)`,
			shouldBe:   false,
			trueValue:  8,
			falseValue: -6,
		},
		{
			str:        `ifExpr(2>1,1==1,1==0)`,
			shouldBe:   true,
			trueValue:  true,
			falseValue: false,
		},
		{
			str:        `ifExpr(2>1,1==1,1==0)`,
			shouldBe:   true,
			trueValue:  true,
			falseValue: false,
		},
		{
			str:        `ifExpr(val("one")==1,pi,0)`,
			shouldBe:   true,
			trueValue:  3.14,
			falseValue: 0,
		},
		{
			str:        `ifExpr(val("one")==2,val("pi"),"")`,
			shouldBe:   false,
			trueValue:  3.14,
			falseValue: "",
		},
		{
			str:        `ifExpr(env("notSet")=="","isEmpty","isFilled")`,
			shouldBe:   true,
			trueValue:  "isEmpty",
			falseValue: "isFilled",
		},
		{
			str:        `ifExpr(min(val("I_AC_L1"),val("I_AC_L2"),val("I_AC_L3"))>0&&avg(val("I_AC_L1"),val("I_AC_L2"),val("I_AC_L3"))>0,1,0)`,
			shouldBe:   false,
			trueValue:  1,
			falseValue: 0,
		},
	}

	for _, x := range testSuite {
		e = New(x.str)
		e.Variables(map[string]interface{}{"one": 1.0, "pi": 3.14})
		_ = e.ParseExpr()
		r := e.Run()
		if x.shouldBe == true {
			if r != x.trueValue {
				t.Errorf("%s Expected %v as output but got %v", x.str, x.trueValue, r)
			}
		} else {
			if r != x.falseValue {
				t.Errorf("%s Expected %v as output but got %v", x.str, x.falseValue, r)
			}
		}
	}

	// Special test case for modbus/tor -> force NaN when environment variable is not set
	var s = `ifExpr(env("PA_AC_LIMIT")=="",float64("NaN"),ifExpr(float64(env("PA_AC_LIMIT"))==0,0,1))`
	e = New(s)
	_ = e.ParseExpr()
	r := e.Run()
	if !math.IsNaN(r.(float64)) {
		t.Errorf("%s Expected %v as output but got %v", s, math.NaN(), r)
	}

}

// Multiply two environment variables
func TestEnvironmentVar(t *testing.T) {

	_ = os.Setenv("x", "3.14")
	_ = os.Setenv("y", "-2")

	e := New(`float64(env("x")) * pow(float64(env("y")),2)`)
	_ = e.ParseExpr()
	result := e.Run()
	if result != 12.56 {
		t.Errorf("Expected 12.56 as output but got %v", result)
	}
}

func TestRegexpMatch(t *testing.T) {
	var ok = map[string]bool{
		`regexpMatch ("^\d+$","1234")`:   true,
		`regexpMatch ("^\d+$",1234)`:     true,
		`regexpMatch ("^[eurt]+$",true)`: true,
		`regexpMatch ("^\d+$","abcd")`:   false,
	}
	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %v from %s as output but got %v", r, s, result)
		}
	}
}

func TestPow(t *testing.T) {
	var ok = map[string]float64{
		`pow(2,0)`:             1,
		`pow(0,2)`:             0.0,
		`pow(2,2)`:             4.0,
		`pow("2","2")`:         4.0,
		`round(pow(2.3,2),2)`:  5.29,
		`round(pow(2,-2.3),2)`: 0.20,
	}
	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			// NaN  must be compared different - check this here
			if math.IsNaN(result.(float64)) && math.IsNaN(r) {
				continue
			}
			t.Errorf("Expected %f from %s as output but got %v", r, s, result)
		}
	}

	// WN: added special case here from the comment in eval.go which
	// combines float and int from variables
	var variables = map[string]interface{}{
		"pi": 3.14159,
		"r":  120,
	}
	s := `round(pow(val("r"),2) * val("pi"),0)`
	e := New(s)
	if e.ParseExpr() == nil {
		e.Variables(variables)
		result := e.Run()
		if result != 45239.0 {
			t.Errorf("Expected 45239 from %s as output but got %v", s, result)
		}
	}

}

func TestIntCast(t *testing.T) {
	_ = os.Setenv("x", "7")
	vars := map[string]interface{}{
		"n":  -1.2,
		"pi": 3.141,
	}
	var ok = map[string]int{
		`int("0")`:      0,
		`int("-1")`:     -1,
		`int(-3.141)`:   -3,
		`int(0)`:        0,
		`int(1)`:        1,
		`int(-1)`:       -1,
		`int(1.9)`:      1,
		`int(-1.9)`:     -1,
		`int(100/3)`:    33,
		`int(true)`:     1,
		`int(false)`:    0,
		`int("-6.7")`:   -6,
		`int(n)`:        -1,
		`int(pi)`:       3,
		`int(val("n"))`: -1,
		`int(env("x"))`: 7,
	}
	for s, r := range ok {
		e := New(s).Variables(vars)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %d from %s as output but got %v", r, s, result)
		}
	}

	var wrong = map[string]float64{
		`int()`:       math.NaN(),
		`int("x")`:    math.NaN(),
		`int("true")`: math.NaN(),
	}

	for s, r := range wrong {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if !math.IsNaN(result.(float64)) {
			t.Errorf("Expected %f from %s as output but got %v", r, s, result)
		}
	}

}

// float64 casting
func TestFloat64Cast(t *testing.T) {

	_ = os.Setenv("x", "1.2")
	var ok = map[string]float64{
		`float64(1>0 && 2>1)`:  1.0,
		`float64(1>2 || 2>3)`:  0.0,
		`float64(pow(2,2))`:    4.0,
		`float64(abs(-1)>0)`:   1.0,
		`float64(1)`:           1.0,
		`float64(0>0)`:         0.0,
		`float64(0)`:           0.0,
		`float64(3.14)`:        3.14,
		"float64(true)":        1.0,
		"float64(false)":       0.0,
		"float64(\"-2.27\")":   -2.27,
		"float64(env(\"x\"))":  1.2,
		"float64(\"NaHallo\")": math.NaN(),
		"float64(\"NaN\")":     math.NaN(), // this is important -> we can force math.NaN
	}

	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			// NaN  must be compared different - check this here
			if math.IsNaN(result.(float64)) && math.IsNaN(r) {
				continue
			}
			t.Errorf("Expected %f from %s as output but got %v", r, s, result)
		}
	}

	// special test case - store unix epoch time into variables and get it out
	// tests $SYS/time/startepoch variable
	now := time.Now().Unix()
	e := New("val(\"epoch\")")
	_ = e.ParseExpr()
	e.Variables(map[string]interface{}{
		"epoch": now,
	})
	result := e.Run()
	if result != now {
		t.Errorf("Expected %v from %s as output but got %v", now, "val(\"epoch\")", result)
	}
}

// round
func TestRound(t *testing.T) {

	var ok = map[string]float64{
		`round(3.14159,3)`:   3.142,
		`round(3.14159,2)`:   3.14,
		`round(3.14159,0)`:   3,
		`round(3.14159,-1)`:  0,
		`round("3.14159",3)`: 3.142,
	}

	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %f from %s as output but got %v", r, s, result)
		}
	}
}

// time ("","")
func TestTime(t *testing.T) {

	var epoch = []string{
		`time("","")`,
		`time("","epoch")`,
		`time("now","epoch")`,
		`time("now","")`,
	}

	// e.g. 1593668389
	for _, s := range epoch {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if x, ok := result.(int64); !ok {
			if x <= 1593667772 {
				t.Errorf("Expected epoch time from %s as output but got %v (%v)", s, result, x)
			}
		}
	}

	var rfc3339 = []string{
		`time("","rfc3339")`,
		`time("","rfc3339")`,
		`time("epoch","rfc3339")`,
		`time("epoch","rfc3339")`,
	}

	// e.g. 2020-07-02T07:39:10+02:00
	for _, s := range rfc3339 {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if x, ok := result.(string); !ok {
			if len(x) <= 19 {
				t.Errorf("Expected rfc3339 time from %s as output but got %v (%v)", s, result, x)
			}
		}
	}

}

// sqrt
func TestSqrt(t *testing.T) {

	var ok = map[string]float64{
		`sqrt(16)`:         4,
		`sqrt("16")`:       4,
		`round(sqrt(3),2)`: 1.73,
	}

	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %f from %s as output but got %v", r, s, result)
		}
	}
}

// val -> an unknown variable must be math.NaN !
func TestVal(t *testing.T) {
	// x is not set - so expect math.NaN
	e := New("val(\"x\")")
	_ = e.ParseExpr()
	result := e.Run()
	if result != "" {
		t.Errorf("%v should be math.NaN", result)
	}
}

// setVal
func TestSetVal(t *testing.T) {

	var ok = map[string]interface{}{
		`setVal("a",true) ; val("a")`:                                  true,
		`setVal("a",false) ; val("a")`:                                 false,
		`setVal("a",0) ; val("a")`:                                     0,
		`setVal("n",10) ; setVal("n",val("n")+3*4) ; val("n")`:         22,
		`setVal("a",int(-3.141)) ; a)`:                                 -3,
		`setVal("a",-3.141) ; val("a")`:                                -3.141,
		`setVal("s","str") ; val("s")`:                                 "str",
		`setVal("s","") ; val("s")`:                                    "",
		`setVal("a",10,"$SYS/b",20) ; val("$SYS/b")`:                   20,
		`setVal("a",10,"$SYS/b",20) ; val("a")`:                        10,
		`setVal("a","1","$SYS/b","20") ; val("$SYS/b")`:                "20",
		`setVal("a","1","$SYS/b","20") ; val("a")`:                     "1",
		`setVal("Text",sprintf("x %.2f y",100/3) ; val("Text")`:        "x 33.33 y",
		`setVal("Text",sprintf("x %.2f y",3/100) ; val("Text")`:        "x 0.03 y",
		`setVal("Text",sprintf("x %.2f y",1000*0.00333) ; Text)`:       "x 3.33 y",
		`setVal("Text",sprintf("x %.2f y",0.00333*1000) ; val("Text")`: "x 3.33 y",
		`setVal("Text",sprintf("x %.2f y",0.0333+1000) ; val("Text")`:  "x 1000.03 y",
		`setVal("Text",sprintf("x %.2f y",1000+0.0333) ; val("Text")`:  "x 1000.03 y",
		`setVal("Text",sprintf("x %.2f y",0.0333-1000) ; val("Text")`:  "x -999.97 y",
		`setVal("Text",sprintf("x %.2f y",1000-0.0333) ; val("Text")`:  "x 999.97 y",
	}

	for k, v := range ok {
		fields := strings.Split(k, " ; ")
		e := New("")
		for _, x := range fields {
			e.SetInput(x)
			_ = e.ParseExpr()
			vRet := e.Run()
			if vRet == nil {
				continue
			}
			if vRet != v {
				t.Errorf("%s failed expected %v and got %v", k, v, vRet)
			}

		}
	}

}

// val -> an unknown variable must be math.NaN !
func TestAvgMaxMin(t *testing.T) {

	var ok = map[string]float64{
		`avg()`:                             math.NaN(),
		`max()`:                             math.NaN(),
		`min()`:                             math.NaN(),
		`avg("x")`:                          math.NaN(),
		`max("x")`:                          math.NaN(),
		`min("x")`:                          math.NaN(),
		`min(1/0)`:                          math.Inf(1),
		`avg(-12.13)`:                       -12.13,
		`max(0,-3.33,97.77)`:                97.77,
		`min(0,-3.33,97.77)`:                -3.33,
		`max(109.5)`:                        109.5,
		`min(10)`:                           10.0,
		`avg(10,20)`:                        15.0,
		`avg(30,"10","20.0","John Doe")`:    20.0,
		`max(10,20)`:                        20.0,
		`min(10,20)`:                        10.0,
		`avg(max(10,20,30),min(""))`:        math.NaN(),
		`avg(max(10,20,30),min(-1.2,-2.4))`: 13.8,
		`min(max(10,20,30),min(-1.2,-2.4))`: -2.4,
		`max(max(10,20,30),min(-1.2,-2.4))`: 30.0,
		`avg(10,20,30,-1.2)`:                14.7,
		`max(10,20,30,-1.2)`:                30.0,
		`min(10,20,30,-1.2)`:                -1.2,
		`avg(10,20,30,-1.2,"ignore wrong strings")`: 14.7,
		`max(10,20,30,-1.2,"ignore wrong strings")`: 30,
		`min(10,20,30,-1.2,"ignore wrong strings")`: -1.2,
		`max("1","2")`:       2,
		`min("1.34","2")`:    1.34,
		`max("10",8)`:        10.0,
		`min("10","8","-6")`: -6,
		`avg("10","8","-6")`: 4,
	}

	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if math.IsNaN(r) && math.IsNaN(result.(float64)) {
			continue
		}
		if result != r {
			t.Errorf("Expected %f from %s as output but got %v", r, s, result)
		}
	}

}

// substr
func TestSubstr(t *testing.T) {
	var ok = map[string]string{
		`substr("",0,0)`:                        "",
		`substr("Hallo",0,0)`:                   "",
		`substr("",2,2)`:                        "",
		`substr("MyNameIsJohn",0,-1)`:           "MyNameIsJohn",
		`substr("MyNameIsJohn",2,-1)`:           "NameIsJohn",
		`substr("MyNameIsJohn",100,-1)`:         "",
		`substr("MyNameIsJohn",2,-100)`:         "",
		`substr("MyNameIsJohn",-4,-1)`:          "John",
		`substr("MyNameIsJohn",-4,3)`:           "Joh",
		`substr("MyNameIsJohn",-4,4)`:           "John",
		`substr("MyNameIsJohn",-4,5)`:           "John",
		`substr("MyNameIsJohn",2,4)`:            "Name",
		`substr("MyNameIsJohn",0,1)`:            "M",
		`substr("MyNameIsJohn",11,1)`:           "n",
		`substr("MyNameIsJohn",12,1)`:           "",
		`substr("MyNameIsJohn",0,12)`:           "MyNameIsJohn",
		`substr("43c9666743c8e667436800",16,8)`: "436800",
	}

	for s, r := range ok {
		e := New(s)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %s from %s as output but got %v", r, s, result)
		}
	}

}

func TestSprintf(t *testing.T) {

	var vars = map[string]interface{}{
		"h":  "srv.demo.at",
		"n":  -15,
		"pi": 3.141,
		"b":  true,
		"i":  255,
	}
	var ok = map[string]string{
		`sprintf("")`:            "",
		`sprintf("a","b")`:       "a%!(EXTRA string=\"b\")",
		`sprintf("%.2f",1/(9/3)`: "0.33",
		`sprintf("%s,%d,%.3f,%t",val("h"),val("n"),val("pi"),b)`: "srv.demo.at,-15,3.141,true",
		`sprintf("%s,%d,%.3f,%t",h,n,pi,b)`:                      "srv.demo.at,-15,3.141,true",
		`sprintf("%x",int(i)`:                                    "ff",
	}
	for s, r := range ok {
		e := New(s).Variables(vars)
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %s from %s as output but got %v", r, s, result)
		}
	}
}

//// register
//func TestRegister(t *testing.T) {
//	var ok = map[string]string{
//		`register("2abc556d80ab",1,2)`:    "556d80ab",
//		`register("",0,0)`:                "",
//		`register("Hallo",0,0)`:           "",
//		`register("",2,2)`:                "",
//		`register("MyNameIsJohn",0,-1)`:   "",
//		`register("MyNameIsJohn",2,-1)`:   "",
//		`register("MyNameIsJohn",100,-1)`: "",
//		`register("MyNameIsJohn",2,-100)`: "",
//		`register("MyNameIsJohn",-4,-1)`:  "",
//		`register("MyNameIsJohn",-4,3)`:   "",
//		`register("MyNameIsJohn",-4,4)`:   "",
//		`register("MyNameIsJohn",-4,5)`:   "",
//		`register("MyNameIsJohn",0,1)`:    "MyNa",
//		`register("MyNameIsJohn",1,2)`:    "meIsJohn",
//		`register("MyNameIsJohn",7,17)`:   "",
//	}
//
//	for s, r := range ok {
//		e := New(s)
//		_ = e.ParseExpr()
//		result := e.Run()
//		if result != r {
//			t.Errorf("Expected %s from %s as output but got %v", r, s, result)
//		}
//	}
//
//}

func TestIsBetween(t *testing.T) {

	_ = os.Setenv("x", "50.5")
	var ok = map[string]bool{
		`isBetween(-1,0,1)`:                               false,
		`isBetween(-1,0,0)`:                               false,
		`isBetween(1,0,1)`:                                true,
		`isBetween("1",0,1)`:                              true,
		`isBetween("1","0","1")`:                          true,
		`isBetween("1",0,0)`:                              false,
		`isBetween(env("x"),0,100)`:                       true,
		`isBetween(env("x"),0,50.5)`:                      true,
		`isBetween(env("x"),50.5,50.5)`:                   true,
		`isBetween(env("x"),50.5,0)`:                      false,
		`isBetween(env("y"),0,100)`:                       false,
		`isBetween(env("x"),val("a"),abs(val("b"))`:       true,
		`isBetween(time("now",""),0,9999999999)`:          false,
		`isBetween(float64(time("now","")),0,9999999999)`: true,
		`isBetween(-0.95,-0.99,-0.90)`:                    true,
		`isBetween(-0.89,-0.99,-0.90)`:                    false,
		`isBetween(something,"Wrong",/)`:                  false,
	}

	for s, r := range ok {
		e := New(s)
		e.Variables(map[string]interface{}{
			"a": 10.7,
			"b": -100.3,
		})
		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %v from %s as output but got %v", r, s, result)
		}
	}
}

func TestIsNaN(t *testing.T) {
	var ok = map[string]bool{
		`isNaN(float64(NaN))`:               true,
		`isNaN(float64(5.5))`:               false,
		`isNaN(5.1)`:                        false,
		`isNaN(555)`:                        false,
		`isNaN(blabla)`:                     true,
		`isNaN("text")`:                     true,
		`isNaN(1>1)`:                        false,
		`isNaN(1==1)`:                       false,
		`isNaN(substr("MyNameIsJohn",2,4))`: true,
		`isNaN(substr("123456.6666",2,7))`:  false,
		`isNaN(   time("now","epoch")  ) `:  false,
		`isNaN(time("now","RFC3339")  ) `:   true,
	}

	for s, r := range ok {
		e := New(s)

		_ = e.ParseExpr()
		result := e.Run()
		if result != r {
			t.Errorf("Expected %v from %s as output but got %v", r, s, result)
		}
	}
}
