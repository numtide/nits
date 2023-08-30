// Code generated by "enumer -type=Action -output=action.go -json"; DO NOT EDIT.

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _ActionName = "ExecuteCancel"

var _ActionIndex = [...]uint8{0, 7, 13}

const _ActionLowerName = "executecancel"

func (i Action) String() string {
	if i < 0 || i >= Action(len(_ActionIndex)-1) {
		return fmt.Sprintf("Action(%d)", i)
	}
	return _ActionName[_ActionIndex[i]:_ActionIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _ActionNoOp() {
	var x [1]struct{}
	_ = x[Execute-(0)]
	_ = x[Cancel-(1)]
}

var _ActionValues = []Action{Execute, Cancel}

var _ActionNameToValueMap = map[string]Action{
	_ActionName[0:7]:       Execute,
	_ActionLowerName[0:7]:  Execute,
	_ActionName[7:13]:      Cancel,
	_ActionLowerName[7:13]: Cancel,
}

var _ActionNames = []string{
	_ActionName[0:7],
	_ActionName[7:13],
}

// ActionString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ActionString(s string) (Action, error) {
	if val, ok := _ActionNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _ActionNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Action values", s)
}

// ActionValues returns all values of the enum
func ActionValues() []Action {
	return _ActionValues
}

// ActionStrings returns a slice of all String values of the enum
func ActionStrings() []string {
	strs := make([]string, len(_ActionNames))
	copy(strs, _ActionNames)
	return strs
}

// IsAAction returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Action) IsAAction() bool {
	for _, v := range _ActionValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for Action
func (i Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for Action
func (i *Action) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Action should be a string, got %s", data)
	}

	var err error
	*i, err = ActionString(s)
	return err
}