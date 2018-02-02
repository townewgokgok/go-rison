package rison

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

var testCases = map[string]string{

	// quoted strings
	"''":                  `""`,
	"'0a'":                `"0a"`,
	"'abc def'":           `"abc def"`,
	"'-h'":                `"-h"`,
	"'user@domain.com'":   `"user@domain.com"`,
	"'US $10'":            `"US $10"`,
	"'wow!!'":             `"wow!"`,
	"'can!'t'":            `"can't"`,
	"'Control-F: \u0006'": `"Control-F: \u0006"`,
	"'Unicode: ௫'":        `"Unicode: ௫"`,

	// bare strings
	"G.":         `"G."`,
	"a":          `"a"`,
	"a-z":        `"a-z"`,
	"domain.com": `"domain.com"`,

	// numbers
	"0":     `0`,
	"1":     `1`,
	"42":    `42`,
	"1.5":   `1.5`,
	"99.99": `99.99`,
	"-3":    `-3`,
	"-33":   `-33`,
	"1e30":  `1e+30`,
	"1e-30": `1e-30`,
	//"1E30":  `1e+30`,
	"1.5e2": `150`,

	// other primitives
	"!t": `true`,
	"!f": `false`,
	"!n": `null`,

	// objects
	"()":                            `{}`,
	"(a:0)":                         `{"a":0}`,
	"(a:0,b:1)":                     `{"a":0,"b":1}`,
	"(a:0,b:foo,c:'23skidoo')":      `{"a":0,"b":"foo","c":"23skidoo"}`,
	"(id:!n,type:/common/document)": `{"id":null,"type":"/common/document"}`,
	`(any:json,yes:!t)`:             `{"any":"json","yes":true}`,

	// arrays
	"!()":            `[]`,
	"!(1,2,3)":       `[1,2,3]`,
	"!(foo,bar)":     `["foo","bar"]`,
	"!(!t,!f,!n,'')": `[true,false,null,""]`,

	// complex objects
	`(A:(B:(C:(D:E,F:G)),H:(I:(J:K,L:M))))`:              `{"A":{"B":{"C":{"D":"E","F":"G"}},"H":{"I":{"J":"K","L":"M"}}}}`,
	`!(A,B,(supportsObjects:!t))`:                        `["A","B",{"supportsObjects":true}]`,
	"(foo:bar,baz:!(1,12e40,0.42,(a:!t,'0':!f,'1':!n)))": `{"foo":"bar","baz":[1,12e40,0.42,{"a":true,"0":false,"1":null}]}`,
}

var invalidCases = []string{

	// objects
	"(",
	"(foo",
	"(foo:",
	"(foo:1",
	")",
	"())",
	"(,",
	"(,)",
	"(foo:1,)",
	"(,bar:2)",
	"(1not:'id')",

	// strings
	"'",
	"'abc",
	"'a!'!'",

	// numbers
	"4abc",
	"-",
	"-h",
	"-1h",
	"--1",
	"1-",
	"-1-",
	"-1-1",
	"1e-",
	"1e-h",
	"1e-1h",
	"1e--1",
	"1e1-",
	"1e-1-",
	"1e-1-1",
	"1.5e+2",
	"1.5E2",
	"1.5E+2",
	"1.5E-2",

	// escape sequences
	"!",
	"!z",
	"!!!",
	"!tf",

	// spaces
	"   ",
	"foo bar",

	// others
	"",
	"!(!t!f)",
	"(a:!t,0:!f,1:!n)",
}

func dumpValue(v interface{}) string {
	j, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(j)
}

func TestDecodeEncode(t *testing.T) {
	for r, j := range testCases {
		var object interface{}
		err := json.Unmarshal([]byte(j), &object)
		if err != nil {
			t.Fatal(err)
		}

		decoded, err := Decode([]byte(r), Mode_Rison)
		if err != nil {
			t.Errorf("decoding %s : want %s, got error `%s`", r, j, err.Error())
		} else if !reflect.DeepEqual(object, decoded) {
			t.Errorf("decoding %s : want %s, got %s", r, j, dumpValue(decoded))
		}

		encoded, err := Encode(object, Mode_Rison)
		if err != nil {
			t.Errorf("encoding %s : want %s, got error `%s`", j, r, err.Error())
		} else {
			redecoded, err := Decode(encoded, Mode_Rison)
			if err != nil {
				t.Errorf("encoding %s : want %s, got %s and error `%s`", j, r, string(encoded), err.Error())
			} else if !reflect.DeepEqual(object, redecoded) {
				t.Errorf("encoding %s : want %s, got %s", j, r, string(encoded))
			}
		}
	}
}

func TestDecodeORison(t *testing.T) {
	r := `a:1,b:!f`
	j := `{"a":1,"b":false}`
	var object interface{}
	err := json.Unmarshal([]byte(j), &object)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode([]byte(r), Mode_ORison)
	if err != nil {
		t.Errorf("decoding %s : want %s, got error `%s`", r, j, err.Error())
	} else if !reflect.DeepEqual(object, decoded) {
		t.Errorf("decoding %s : want %s, got %s", r, j, dumpValue(decoded))
	}
}

func TestDecodeARison(t *testing.T) {
	r := `a,2,!t`
	j := `["a",2,true]`
	var object interface{}
	err := json.Unmarshal([]byte(j), &object)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode([]byte(r), Mode_ARison)
	if err != nil {
		t.Errorf("decoding %s : want %s, got error `%s`", r, j, err.Error())
	} else if !reflect.DeepEqual(object, decoded) {
		t.Errorf("decoding %s : want %s, got %s", r, j, dumpValue(decoded))
	}
}

func TestDecodeDeepNestedObject(t *testing.T) {
	l := ""
	r := ""
	for i := 0; i < 10000; i++ {
		l += "(a:1,b:"
		r += ",c:3)"
	}
	l += "2"
	_, err := Decode([]byte(l+r), Mode_Rison)
	if err != nil {
		t.Errorf("decoding %s .. : want no error, got error `%s`", l[:100], err.Error())
	}
}

func TestDecodeDeepNestedArray(t *testing.T) {
	l := ""
	r := ""
	for i := 0; i < 10000; i++ {
		l += "!(!(),"
		r += ",!())"
	}
	l += "!()"
	_, err := Decode([]byte(l+r), Mode_Rison)
	if err != nil {
		t.Errorf("decoding %s .. : want no error, got error `%s`", l[:100], err.Error())
	}
}

func TestDecodeErrors(t *testing.T) {
	for _, r := range invalidCases {
		decoded, err := Decode([]byte(r), Mode_Rison)
		if err == nil {
			t.Errorf("decoding %s : want an error, got %s", r, dumpValue(decoded))
		}
	}
}