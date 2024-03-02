package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// readJsonStdin is a utility function that reads JSON from stdin
// and returns an any
func readJsonStdin() (any, error) {
	var data any
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("cannot read JSON input: %w", err)
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal JSON data: %w", err)
	}
	return data, nil
}

// getKAny is a utility function that type casts an any
// and returns a map of string and any
// if the input is neither one of these we will return nil
func getKAny(o any) map[string]any {
	if val, ok := o.(map[string]any); ok {
		return val
	}
	if val, ok := o.([]any); ok {
		arr := make(map[string]any)
		for i, v := range val {
			arr[fmt.Sprintf("%d", i)] = v
		}
		return arr
	}
	return nil
}

// getVal is a utility function that takes any and returns
// an appropriate string value for it
func getVal(o any) string {
	if valstr, ok := o.(string); ok {
		return valstr
	}
	if valint, ok := o.(int); ok {
		return fmt.Sprintf("%d", valint)
	}
	if valflt, ok := o.(float64); ok {
		return fmt.Sprintf("%f", valflt)
	}
	if valbool, ok := o.(bool); ok {
		return fmt.Sprintf("%t", valbool)
	}
	if _, ok := o.(map[string]any); ok {
		return "{}"
	}
	if _, ok := o.([]any); ok {
		return "[]"
	}
	return ""
}

// getInitialKV is a utility function that gets the initial list of key-value pairs
// given an any
func getInitialKV(o any) []KVPair {
	kvpairs := []KVPair{}
	m := getKAny(o)
	if m != nil {
		for key, val := range m {
			kvp := KVPair{
				Key:   key,
				Value: getVal(val),
			}
			kvpairs = append(kvpairs, kvp)
		}
	}
	return kvpairs
}
