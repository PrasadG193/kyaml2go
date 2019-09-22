package stack

type Stack interface {
	Push(interface{})
	Pop() (interface{}, bool)
	Top() (interface{}, bool)
	Length() int
}

type DefaultStack struct {
	Values []interface{}
}

func New() Stack {
	return &DefaultStack{}
}

func (s *DefaultStack) Push(value interface{}) {
	s.Values = append(s.Values, value)
}

func (s *DefaultStack) Pop() (value interface{}, exists bool) {
	if len(s.Values) == 0 {
		return value, false
	}
	item := s.Values[len(s.Values)-1]
	s.Values = s.Values[:len(s.Values)-1]
	return item, true
}

func (s DefaultStack) Top() (value interface{}, exist bool) {
	if len(s.Values) == 0 {
		return value, exist
	}
	return s.Values[len(s.Values)-1], true
}

func (s DefaultStack) Length() int {
	return len(s.Values)
}
