package main

import "slices"

type TokenArrayInterface interface {
	at(index uint) Token
	push(tokens ...Token) uint
}

type TokenArray struct {
	elements []Token
	length   uint
}

// returns the token at the specified index
func (a *TokenArray) at(index uint) Token {
	return a.elements[index]
}

// adds the new tokens to the end of the array
// and returns the new length of the array
func (a *TokenArray) push(tokens ...Token) uint {
	a.elements = append(a.elements, tokens...)
	newLength := uint(len(a.elements))
	a.length = newLength
	return newLength
}

func tokenArray(tokens ...Token) *TokenArray {
	elements := tokens
	length := len(elements)
	return &TokenArray{elements, uint(length)}
}

// type ArrayStruct struct{}

// func (p *ArrayStruct) contains(array []any, element any) bool {
// 	found := false
// 	for _, el := range array {
// 		if el == element {
// 			found = true
// 			break
// 		}
// 	}
// 	return found
// }

// var Array = ArrayStruct{}
//#region CallStack

type Stack struct {
	stack  []FunctionVal
	length int
}

func (s *Stack) at(index int) *FunctionVal {
	if index < 0 {
		index += s.length
	}
	v := s.stack[index]
	return &v
}

func (s *Stack) Push(elements ...FunctionVal) *Stack {
	length := &s.length
	// do not use range over loop
	for i := 0; i < len(elements); i++ {
		el := elements[i]
		s.stack = append(s.stack, el)
		index := i + *length
		*length++
		if index >= *length {
			s.length = index + 1
		}
	}
	return s
}

func (s *Stack) Pop() FunctionVal {
	if s.length == 0 {
		var nilFn FunctionVal
		return nilFn
	}
	return s.Delete(-1)
}

func (s *Stack) Delete(index int) FunctionVal {
	if index < 0 {
		index += s.length
	}
	v := s.stack[index]
	s.stack = slices.Delete(s.stack, index, index+1)
	s.length = len(s.stack)
	return v
}

func NewStack() *Stack {
	return &Stack{
		stack:  []FunctionVal{},
		length: 0,
	}
}
