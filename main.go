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

// readJsonStdin is a utility function that reads JSON from stdin
// and returns a map of keys as strings and values as interface{}
func readJsonStdin() (map[string]interface{}, error) {
	var data map[string]interface{}
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

// getInitialMap gets the first list of key-value pairs in the JSON
func getInitialMap(data map[string]interface{}) map[string]string {
	startMap := make(map[string]string)
	for key, val := range data {
		if valstr, ok := val.(string); ok {
			startMap[key] = valstr
		} else {
			startMap[key] = "{}" // I'll accomodate lists later
		}
	}
	return startMap
}

// getInitialKeys gets the first list of keys in the JSON
func getInitialKeys(data map[string]interface{}) []string {
	keys := []string{}
	for key, _ := range data {
		keys = append(keys, key)
	}
	return keys
}

// Cursor contains the cursor's horizontal and vertical position
type Cursor struct {
	selKeys   []string // keys selected
	currSel   string   // current key selection
	currKeys  []string // current keys in current map
	currIndex int      // current index of current keys
}

// Model contains the data and its visual representation
// data holds the actual parsed data
// cursorVertical holds the position of the cursor on the current list
// cursorHorizontal holds the position of the cursor at the current JSON object
type Model struct {
	data    map[string]interface{}
	currMap map[string]string
	cursor  Cursor
	isLeaf  bool
}

// NewModel creates a new Model struct with initial settings
func NewModel() *Model {
	// we will read the JSON from Stdin
	data, err := readJsonStdin()
	if err != nil {
		return nil
	}
	c := Cursor{
		currKeys: getInitialKeys(data),
		currSel:  "",
	}
	return &Model{
		data:    data,
		currMap: getInitialMap(data),
		cursor:  c,
		isLeaf:  false,
	}
}

// updateCurrentMap updates the current map depending on what keys are
// selected
func (m *Model) updateCurrentMap() {
	// something needs to be selected
	// and the value of the last selected must be a map
	if len(m.cursor.selKeys) > 0 && !m.isLeaf {
		// get the first map
		tempMap, _ := m.data[m.cursor.selKeys[0]].(map[string]interface{})
		// iterate over the selected keys
		for i := 1; i < len(m.cursor.selKeys)-1; i++ {
			tempMap, _ = tempMap[m.cursor.selKeys[i]].(map[string]interface{})
		}
		// reset the current map
		m.currMap = make(map[string]string)
		for key, val := range tempMap {
			if valstr, ok := val.(string); ok {
				m.currMap[key] = valstr
			} else {
				m.currMap[key] = "{}" // I'll accomodate lists later
			}
		}
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor.currIndex > 0 {
				m.cursor.currIndex--
			}
			// change current selection to the one the cursor is pointing to
			if len(m.cursor.selKeys) > 0 && m.cursor.currSel == m.cursor.selKeys[len(m.cursor.selKeys)-1] {
				m.cursor.selKeys = m.cursor.selKeys[:len(m.cursor.selKeys)-1]
				m.cursor.currSel = ""
				m.isLeaf = false
			}
		case "down":
			if m.cursor.currIndex < len(m.cursor.currKeys)-1 {
				m.cursor.currIndex++
			}
			// change current selection to the one the cursor is pointing to
			if len(m.cursor.selKeys) > 0 && m.cursor.currSel == m.cursor.selKeys[len(m.cursor.selKeys)-1] {
				m.cursor.selKeys = m.cursor.selKeys[:len(m.cursor.selKeys)-1]
				m.cursor.currSel = ""
				m.isLeaf = false
			}
		case "right":
			// right moves cursor from key to value
			// the key at the cursor is selected
			m.cursor.currSel = m.cursor.currKeys[m.cursor.currIndex]
			// check if the value of the key is another map
			if val := m.currMap[m.cursor.currSel]; val != "{}" {
				// we are at the bottom of the tree
				m.isLeaf = true
			}
		case "left":
			// left moves cursor from value to key
			// we will no longer be at a leaf
			if m.isLeaf == true {
				m.isLeaf = false
			}
			m.cursor.currSel = ""

		case "enter":
			// if we are not at a leaf and there is a selection
			if m.isLeaf == false && m.cursor.currSel != "" {
				// update the current map to the one at the last selected key
				m.cursor.selKeys = append(m.cursor.selKeys, m.cursor.currSel)
				m.updateCurrentMap()
				// reset the current keys
				m.cursor.currKeys = []string{}
				for key, _ := range m.currMap {
					m.cursor.currKeys = append(m.cursor.currKeys, key)
				}
				// reset the current key index
				m.cursor.currIndex = 0
				// reset the current selected key
				m.cursor.currSel = ""
			}
		}
	}
	return m, nil
}

func (m *Model) View() string {
	s := ""
	if len(m.cursor.selKeys) > 0 {
		for _, sel := range m.cursor.selKeys {
			s += fmt.Sprintf("%s:", sel)
		}
		s += fmt.Sprintf("\n\n")
	}
	for index, key := range m.cursor.currKeys {
		cursor := " "
		if m.cursor.currIndex == index {
			cursor = ">"
			if m.isLeaf {
				cursor = "<"
			}
		}
		if m.cursor.currSel != "" {
			s += fmt.Sprintf("%s: %s %s\n", key, cursor, m.currMap[key])
		} else {
			s += fmt.Sprintf("%s %s: %s\n", cursor, key, m.currMap[key])
		}
	}
	s += "\nQuit: ctrl+c  Up: ↑  Down: ↓  Left: ←  Right: → \n"
	return s
}
