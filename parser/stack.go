package parser

type Stack struct {
	elements []interface{}
}

// NewStack creates a new stack instance
func NewStack() *Stack {
	return &Stack{
		elements: make([]interface{}, 0),
	}
}

// Core operations:

// Push Data onto the stack:
func (s *Stack) Push(data interface{}) {
	s.elements = append(s.elements, data)
}

// Pop Data from the stack:
func (s *Stack) Pop() interface{} {
	if len(s.elements) == 0 {
		return nil
	}

	n := len(s.elements) - 1
	data := s.elements[n]
	// this next line is optional, but it's safer to implement to avoid memory leaks
	s.elements[n] = nil
	s.elements = s.elements[:n]
	return data
}

// Peek at the top element of the stack without removing it:
func (s *Stack) Peek() interface{} {
	if len(s.elements) == 0 {
		return nil
	}
	return s.elements[len(s.elements)-1]
}

// IsEmpty checks if the stack is empty
func (s *Stack) IsEmpty() bool {
	return len(s.elements) == 0
}

// Size returns the number of elements in the stack
func (s *Stack) Size() int {
	return len(s.elements)
}
