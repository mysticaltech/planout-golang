/*
 * Copyright 2015 URX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package planout

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func runExperiment(rawCode []byte) (*Interpreter, bool) {

	code := make(map[string]interface{})
	json.Unmarshal(rawCode, &code)

	// fmt.Printf("Code: %v\n", code)

	expt := &Interpreter{
		Salt:      "test_salt",
		Evaluated: false,
		Inputs:    map[string]interface{}{},
		Outputs:   map[string]interface{}{},
		Overrides: map[string]interface{}{},
		Code:      code,
	}

	_, ok := expt.Run()

	return expt, ok
}

func runConfig(config []byte) (*Interpreter, bool) {
	setX := []byte(`{"op": "set", "var": "x", "value":`)
	end := []byte(`}`)

	rawCode := append(setX, config...)
	rawCode = append(rawCode, end...)

	return runExperiment(rawCode)
}

func TestCoreOps(t *testing.T) {

	// Test SET
	expt, _ := runExperiment([]byte(`{"op": "set", "value": "x_val", "var": "x"}`))
	x, _ := expt.get("x")
	if compare(x, "x_val") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	// Test SEQ
	expt, _ = runExperiment([]byte(`
	 	{"op": "seq",
	 	"seq": [ {"op": "set", "value": "x_val", "var": "x"},
	 		 {"op": "set", "value": "y_val", "var": "y"} ]}`))
	x, _ = expt.get("x")
	if compare(x, "x_val") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}
	y, _ := expt.get("y")
	if compare(y, "y_val") != 0 {
		t.Errorf("Variable y. Expected y_val. Actual %v\n", y)
	}

	// Test Array
	expt, _ = runExperiment([]byte(`
	 	{"op": "set", "var": "x", "value": {"op": "array", "values": [4, 5, "a"]}}`))
	x, _ = expt.get("x")

	// Test Condition
	expt, _ = runExperiment([]byte(`
	 	{"op": "cond",
	 	"cond": [ {"if": 0, "then": {"op": "set", "var": "x", "value": "x_0"}},
	 		  {"if": 1, "then": {"op": "set", "var": "x", "value": "x_1"}}]}`))
	x, _ = expt.get("x")
	if compare(x, "x_1") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	expt, _ = runExperiment([]byte(`
	 	{"op": "cond",
	 	"cond": [ {"if": 1, "then": {"op": "set", "var": "x", "value": "x_0"}},
	 		  {"if": 0, "then": {"op": "set", "var": "x", "value": "x_1"}}]}`))
	x, _ = expt.get("x")
	if compare(x, "x_0") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	// Test GET
	expt, _ = runExperiment([]byte(`
	 	{"op": "seq",
	 	"seq": [{"op": "set", "var": "x", "value": "x_val"},
	 		{"op": "set", "var": "y", "value": {"op": "get", "var": "x"}}]}`))
	x, _ = expt.get("x")
	if compare(x, "x_val") != 0 {
		t.Errorf("Variable x. Expected x_val. Actual %v\n", x)
	}

	y, _ = expt.get("y")
	if compare(y, "x_val") != 0 {
		t.Errorf("Variable y. Expected x_val. Actual %v\n", y)
	}

	// Test Index
	expt, _ = runConfig([]byte(` {"op": "index", "index": 0, "base": [10, 20, 30]}`))
	x, _ = expt.get("x")
	if compare(x, 10) != 0 {
		t.Errorf("Variable x. Expected 10. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": 2, "base": [10, 20, 30]}`))
	x, _ = expt.get("x")
	if compare(x, 30) != 0 {
		t.Errorf("Variable x. Expected 30. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": "a", "base": {"a": 42, "b": 43}}`))
	x, _ = expt.get("x")
	if compare(x, 42) != 0 {
		t.Errorf("Variable x. Expected 42. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": 6, "base": [10, 20, 30]}`))
	x, _ = expt.get("x")
	if x != nil {
		t.Errorf("Variable x. Expected nil. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": "c", "base": {"a": 42, "b": 43}}`))
	x, _ = expt.get("x")
	if x != nil {
		t.Errorf("Variable x. Expected nil. Actual %v\n", x)
	}

	expt, _ = runConfig([]byte(` {"op": "index", "index": 2, "base": {"op": "array", "values": [10, 20, 30]}}`))
	x, _ = expt.get("x")
	if compare(x, 30) != 0 {
		t.Errorf("Variable x. Expected 30. Actual %v\n", x)
	}

	// Test Coalesce
	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": {"op": "array", "values": [100, 200, 300]}}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v of type %v\n", x, reflect.TypeOf(x))

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [100, 200, 300, null]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v of type %v\n", x, reflect.TypeOf(x))

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [null]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v of type %v\n", x, reflect.TypeOf(x))

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [null, 42, null]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v of type %v\n", x, reflect.TypeOf(x))

	expt, _ = runConfig([]byte(`{"op": "coalesce", "values": [null, null, 43]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v of type %v\n", x, reflect.TypeOf(x))

	// Test Length
	expt, _ = runConfig([]byte(`{"op": "length", "values": {"op": "array", "values": [1,2,3,4,5]}}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)

	expt, _ = runConfig([]byte(`{"op": "length", "values": [1,2,3,4,5]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)

	expt, _ = runExperiment([]byte(`{"op":"seq","seq":[{"op":"set","var":"arr","value":{"op":"array","values":[111,222,333]}},{"op":"set","var":"x","value":{"values":[{"op":"get","var":"arr"}],"op":"length"}}]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)

	expt, _ = runExperiment([]byte(`{"op":"seq","seq":[{"op":"set","var":"a","value":111},{"op":"set","var":"b","value":222},{"op":"set","var":"c","value":{"op":"array","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]}},{"op":"set","var":"x","value":{"values":[{"op":"get","var":"c"}],"op":"length"}}]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)

	expt, _ = runExperiment([]byte(`{"op":"seq","seq":[{"op":"set","var":"a","value":111},{"op":"set","var":"b","value":222},{"op":"set","var":"x","value":{"values":[{"op":"array","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]}],"op":"length"}}]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)

	expt, _ = runExperiment([]byte(`{"op":"seq","seq":[{"op":"set","var":"a","value":1111},{"op":"set","var":"x","value":{"values":[{"op":"array","values":[{"op":"get","var":"a"},3333]}],"op":"length"}}]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)

	expt, _ = runExperiment([]byte(`{"op":"seq","seq":[{"op":"set","var":"x","value":{"values":[{"op":"array","values":[111,222]}],"op":"length"}}]}`))
	x, _ = expt.get("x")
	fmt.Printf("X: %v\n", x)
}
