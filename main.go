package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	err := tea.NewProgram(NewModel()).Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// KVPair stores a key and value used in a map
type KVPair struct {
	Key   string
	Value string
}

// Cursor contains the cursor's horizontal and vertical position
type Cursor struct {
	RowNo         int    // used to move the cursor up and down
	IsKey         bool   // used to indicate if the cursor is pointing to a key or value
	IsEnd         bool   // used to indicate if we have come to the end of a path
	CursorDisplay string // this gets displayed to the console
}

// Model contains the data and its visual representation
type Model struct {
	Data   any      // contains the parsed JSON data
	CurrC  Cursor   // the cursor position
	CurrKV []KVPair // current list of key-value pairs
	Path   []string // current path location
}

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

// NewModel gets the initial model
func NewModel() *Model {
	// we will read the JSON from Stdin
	data, err := readJsonStdin()
	if err != nil {
		return nil
	}
	kvpairs := getInitialKV(data)
	// if there are no key-value pairs there is nothing to do
	if len(kvpairs) == 0 {
		return nil
	}
	c := Cursor{
		RowNo:         0,     // first row is always 0
		IsKey:         true,  // first thing the cursor points to is a key
		IsEnd:         false, // this is the very start of the path
		CursorDisplay: "→",   // we go right
	}
	return &Model{
		Data:   data,
		CurrC:  c,
		CurrKV: kvpairs,
		Path:   []string{}, // path is empty in the beginning
	}
}

// TODO: ask for a path to a file if no stdin data
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update updates the model based on tea.KeyMsg
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		// cursor moving up and down changes the RowNo
		// this action means we are moving through keys
		case "up":
			if m.CurrC.RowNo > 0 {
				m.CurrC.RowNo--
			}
			m.CurrC.IsKey = true
			m.CurrC.IsEnd = false
			m.CurrC.CursorDisplay = "→"
		case "down":
			if m.CurrC.RowNo < len(m.CurrKV)-1 {
				m.CurrC.RowNo++
			}
			m.CurrC.IsKey = true
			m.CurrC.IsEnd = false
			m.CurrC.CursorDisplay = "→"
		// left and right keys moves the cursor from key to value
		// if the cursor is at the end of a path it can only go left
		case "right":
			if m.CurrC.IsKey {
				// always pointing at a value
				m.CurrC.IsKey = false
				// Check if this is an end value
				if m.CurrKV[m.CurrC.RowNo].Value != "{}" && m.CurrKV[m.CurrC.RowNo].Value != "[]" {
					m.CurrC.IsEnd = true
					// update CursorDisplay
					m.CurrC.CursorDisplay = "←"
				} else {
					m.CurrC.IsEnd = false
					m.CurrC.CursorDisplay = "→"
				}
			}
		case "left":
			// always pointing at a key
			m.CurrC.IsKey = true
			// no longer at the end
			m.CurrC.IsEnd = false
			m.CurrC.CursorDisplay = "→"

		// enter expands a {} or [] value which turns into a new list of key-value pairs
		// enter does nothing if it is at a key or if it is at a value that cannot expand
		case "enter":
			if !m.CurrC.IsKey && !m.CurrC.IsEnd {
				// append the current Key to the Path
				m.Path = append(m.Path, m.CurrKV[m.CurrC.RowNo].Key)
				// update the model
				m.CurrC.IsKey = true
				m.CurrC.RowNo = 0
				m.CurrC.IsEnd = false
				m.CurrC.CursorDisplay = "→"
				m.updateKV()
			}
		// x goes back one key and reloads the previous key-value pairs
		case "x":
			// remove the last selected key and update the current map
			if len(m.Path) > 0 {
				m.Path = m.Path[:len(m.Path)-1]
			}
			// update the model
			m.CurrC.IsKey = true
			m.CurrC.RowNo = 0
			m.CurrC.IsEnd = false
			m.CurrC.CursorDisplay = "→"
			m.updateKV()
		}
	}
	return m, nil
}

// updateKV updates the model's list of key-value pairs
func (m *Model) updateKV() {
	// remove everything from the current key-value pair list
	m.CurrKV = nil
	// if there is nothing in the path just fill the first set of key-value pairs
	if len(m.Path) == 0 {
		m.CurrKV = getInitialKV(m.Data)
	} else {
		// iterate through the Path to get the final key-pair
		tempMap := getKAny(m.Data)
		if tempMap != nil {
			for _, k := range m.Path {
				o := tempMap[k]      // gets an any object
				tempMap = getKAny(o) // converts it into a map of string and any

			}
			// we now have a key-value pair which we can fill out
			for k, v := range tempMap {
				m.CurrKV = append(m.CurrKV, KVPair{Key: k, Value: getVal(v)})
			}
		}
	}
}

func (m *Model) View() string {
	s := "You are here: "
	if len(m.Path) > 0 {
		for _, p := range m.Path {
			s += fmt.Sprintf("%s: ", p)
		}
	}
	s += fmt.Sprintf("\n\n")
	for index, kv := range m.CurrKV {
		if m.CurrC.RowNo == index {
			if m.CurrC.IsKey {
				s += fmt.Sprintf("%s %s: %s\n", m.CurrC.CursorDisplay, kv.Key, kv.Value)
			} else {
				s += fmt.Sprintf("%s: %s %s\n", kv.Key, m.CurrC.CursorDisplay, kv.Value)
			}
		} else {
			s += fmt.Sprintf("%s: %s\n", kv.Key, kv.Value)
		}
	}
	s += "\nQuit: ctrl+c  Up: ↑  Down: ↓  Left: ←  Right: →  Expand: enter  Back: x \n"
	return s
}
