package main

import (
	"fmt"

	page "github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
)

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
	Data   any        // contains the parsed JSON data
	CurrC  Cursor     // the cursor position
	CurrKV []KVPair   // current list of key-value pairs
	Path   []string   // current path location
	Page   page.Model // paginator
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
	p := page.New()
	// unbind the default key bindings of the paginator
	p.KeyMap.PrevPage.Unbind()
	p.KeyMap.NextPage.Unbind()
	p.SetTotalPages(len(kvpairs))
	return &Model{
		Data:   data,
		CurrC:  c,
		CurrKV: kvpairs,
		Path:   []string{}, // path is empty in the beginning
		Page:   p,
	}
}

// TODO: ask for a path to a file if no stdin data
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update updates the model based on tea.KeyMsg
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Page.PerPage = msg.Height - 5
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
				m.Page.SetTotalPages(len(m.CurrKV))
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
			m.Page.SetTotalPages(len(m.CurrKV))
		}
	}
	m.Page, cmd = m.Page.Update(msg)
	return m, cmd
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

// getPageItems is a utility function that returns the list of key-value pairs in string form
func (m *Model) getPageItems() []string {
	items := []string{}
	for index, kv := range m.CurrKV {
		if m.CurrC.RowNo == index {
			if m.CurrC.IsKey {
				items = append(items, fmt.Sprintf("%s %s: %s", m.CurrC.CursorDisplay, kv.Key, kv.Value))
			} else {
				items = append(items, fmt.Sprintf("%s: %s %s", kv.Key, m.CurrC.CursorDisplay, kv.Value))
			}
		} else {
			items = append(items, fmt.Sprintf("%s: %s", kv.Key, kv.Value))
		}
	}
	return items
}

func (m *Model) View() string {
	s := "You are here: "
	if len(m.Path) > 0 {
		for _, p := range m.Path {
			s += fmt.Sprintf("%s: ", p)
		}
	}
	s += "\n\n"
	items := m.getPageItems()
	start, end := m.Page.GetSliceBounds(len(items))
	if m.CurrC.RowNo < start {
		m.Page.PrevPage()
	}
	if m.CurrC.RowNo > end {
		m.Page.NextPage()
	}
	for _, item := range items[start:end] {
		s += fmt.Sprintf("%s\n", item)
	}
	s += m.Page.View()
	s += "\n\nQuit: ctrl+c  Up: ↑  Down: ↓  Left: ←  Right: →  Expand: enter  Back: x \n"
	return s
}
