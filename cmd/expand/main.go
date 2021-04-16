package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"io"
	"os"
)

func isHexEncoded(field string) bool {
	return len(field) >= 2 && field[0:2] == "\\x"
}

func decode(field string) interface{} {
	var intermediate interface{}
	bytes := []byte(field)

	if isHexEncoded(field) {
		decoded_bytes, err := hex.DecodeString(field[2:])
		if err != nil {
			panic(err)
		}
		err = msgpack.Unmarshal(decoded_bytes, &intermediate)
		if err != nil {
			panic(err)
		}
		return intermediate
	}

	err := json.Unmarshal(bytes, &intermediate)
	if err != nil {
		panic(err)
	}
	return intermediate
}

func encode(data interface{}) string {
	result, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func reencode_field(field string) string {
	if field == "" {
		return field
	}
	return encode(decode(field))
}

func expand_compact_flow(flow interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	switch flow.(type) {
	case []interface{}:
		switch flow.([]interface{})[0] {
		case "S":
			result["class"] = "Dynflow::Flows::Sequence"
		case "C":
			result["class"] = "Dynflow::Flows::Concurrence"
		default:
			panic("Unknown flow type")
		}
		var subflows []interface{}
		for subflow := range flow.([]interface{})[1:] {
			subflows = append(subflows, expand_compact_flow(subflow))
		}
		result["flows"] = subflows
	case float64, int:
		result["class"] = "Dynflow::Flows::Atom"
		result["step_id"] = flow
	default:
		panic("Unknown flow type")
	}
	return result
}

func expand_flow(field string) string {
	if field == "" {
		return field
	}
	intermediate := decode(field)

	var result map[string]interface{}
	switch intermediate.(type) {
	// old style hash
	case map[string]interface{}:
		result = intermediate.(map[string]interface{})

	// newer compact S-expression like representation
	case []interface{}, float64:
		result = expand_compact_flow(intermediate)
	}
	return encode(result)
}

var encoded_step_columns = [...]int{3, 12, 14}

func expand_step(record []string) []string {
	// 0                   1  2         3    4     5          6        7         8              9             10              11    12    13           14       15
	// execution_plan_uuid,id,action_id,data,state,started_at,ended_at,real_time,execution_time,progress_done,progress_weight,class,error,action_class,children,queue
	//
	// encoded columns are:
	// 3 - data
	// 12 - error
	// 14 - children
	for i := range encoded_step_columns {
		record[i] = reencode_field(record[i])
	}
	return record
}

var encoded_action_columns = [...]int{3, 12, 14}

func expand_action(record []string) []string {
	// 0                   1  2    3                        4                5     6     7      8            9           10
	// execution_plan_uuid,id,data,caller_execution_plan_id,caller_action_id,class,input,output,plan_step_id,run_step_id,finalize_step_id
	//
	// encoded columns are:
	// 2 - data
	// 6 - input
	// 7 - output
	for i := range encoded_action_columns {
		record[i] = reencode_field(record[i])
	}
	return record
}

func expand_execution_plan(record []string) []string {
	// Without msgpack
	// 0    1    2     3      4          5        6         7              8     9     10       11            12                13                14
	// uuid,data,state,result,started_at,ended_at,real_time,execution_time,label,class,run_flow,finalize_flow,execution_history,root_plan_step_id,step_ids
	//
	// encoded columns are:
	// 1 - data
	// 10 - run_flow
	// 11 - finalize_flow
	// 12 - execution_history
	// 14 - step_ids

	// With msgpack
	// 0    1    2     3      4          5        6         7              8     9     10                11       12            13                14
	// uuid,data,state,result,started_at,ended_at,real_time,execution_time,label,class,root_plan_step_id,run_flow,finalize_flow,execution_history,step_ids
	//
	// 1 - data
	// 11 - run_flow
	// 12 - finalize_flow
	// 13 - execution_history
	// 14 - step_ids

	var encoded_execution_plan_columns [3]int
	var flow_columns [2]int

	// The step_ids field should be a safe indicator
	if isHexEncoded(record[14]) {
		encoded_execution_plan_columns = [...]int{1, 13, 14}
		flow_columns = [...]int{11, 12}
	} else {
		encoded_execution_plan_columns = [...]int{1, 12, 14}
		flow_columns = [...]int{10, 11}
	}

	for _, i := range encoded_execution_plan_columns {
		record[i] = reencode_field(record[i])
	}
	for _, i := range flow_columns {
		record[i] = expand_flow(record[i])
	}
	return record
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments, expected 1.\n")
		os.Exit(1)
	}

	var expansion_f func([]string) []string

	switch os.Args[1] {
	case "action":
		expansion_f = expand_action
	case "step":
		expansion_f = expand_step
	case "execution_plan":
		expansion_f = expand_execution_plan
	default:
		fmt.Fprintf(os.Stderr, "Unknown argument '%s', expected one of 'action', 'step' or 'execution_plan'\n", os.Args[1])
		os.Exit(1)
	}

	reader := csv.NewReader(os.Stdin)
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		writer.Write(expansion_f(record))
	}
}
