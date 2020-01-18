package stack

// Stack a native implementation for stack data structure
type Stack interface {
	Push(interface{})
	Pop() (interface{}, bool)
	Top() (interface{}, bool)
	Length() int
}

// DefaultStack is a default implementation for Stack
type DefaultStack struct {
	Values []interface{}
}

// New returns an instance of Stack
func New() Stack {
	return &DefaultStack{}
}

// Push a value to the stack
func (s *DefaultStack) Push(value interface{}) {
	s.Values = append(s.Values, value)
}

// Pop value from the stack
func (s *DefaultStack) Pop() (value interface{}, exists bool) {
	if len(s.Values) == 0 {
		return value, false
	}
	item := s.Values[len(s.Values)-1]
	s.Values = s.Values[:len(s.Values)-1]
	return item, true
}

// Top returns top value from the stack
func (s DefaultStack) Top() (value interface{}, exist bool) {
	if len(s.Values) == 0 {
		return value, exist
	}
	return s.Values[len(s.Values)-1], true
}

// Length returns stack size
func (s DefaultStack) Length() int {
	return len(s.Values)
}
