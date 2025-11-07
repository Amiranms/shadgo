//go:build !solution

package main

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrStackEmpty                = errors.New("stack is empty")
	ErrNotEnoughElements         = errors.New("not enough elements in stack")
	ErrInvalidFunctionDefinition = errors.New("invalid function definition")
	ErrDivideByZero              = errors.New("integer divide by zero")
	ErrUnprocessableEntity       = func(token string) error { return errors.New("unprocessable entity: " + token) }
	ErrUnknownToken              = func(token string) error { return errors.New("unknown toke: " + token) }
)

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type stack []int

func (s *stack) Push(v int) {
	*s = append(*s, v)
}

func (s stack) Top() (int, error) {
	if len(s) == 0 {
		return 0, ErrStackEmpty
	}
	return s[len(s)-1], nil
}

func (s *stack) Pop() (int, error) {
	if len(*s) == 0 {
		return 0, ErrStackEmpty
	}
	index := len(*s) - 1
	elem := (*s)[index]
	*s = (*s)[:index]
	return elem, nil
}

func (s *stack) Over() (int, error) {
	if len(*s) > 1 {
		element := (*s)[len(*s)-2]
		s.Push(element)
		return element, nil
	}
	return 0, ErrNotEnoughElements
}

func (s *stack) Swap() error {
	top0, _ := s.Pop()
	top1, err := s.Pop()
	if err != nil {
		return err
	}
	s.Push(top0)
	s.Push(top1)
	return nil
}

func pushConstant(val int) func(s *stack) error {
	return func(s *stack) error {
		s.Push(val)
		return nil
	}
}

type Evaluator struct {
	KnownOperation map[string]func(*stack) error
	Stack          *stack
	Row            []string
}

func (e *Evaluator) ReadWord() string {
	if len(e.Row) == 0 {
		return ""
	}
	word := e.Row[0]
	e.Row = e.Row[1:]
	return strings.ToLower(word)
}

func NewEvaluator() *Evaluator {
	NewEvaluator := Evaluator{Stack: new(stack)}
	knownOperation := map[string]func(*stack) error{
		"+": func(s *stack) error {
			a, _ := s.Pop()
			b, err := s.Pop()
			if err != nil {
				return err
			}
			s.Push(b + a)
			return nil
		},
		"-": func(s *stack) error {
			a, _ := s.Pop()
			b, err := s.Pop()
			if err != nil {
				return err
			}
			s.Push(b - a)
			return nil
		},
		"*": func(s *stack) error {
			a, _ := s.Pop()
			b, err := s.Pop()
			if err != nil {
				return err
			}
			s.Push(b * a)
			return nil
		},
		"/": func(s *stack) error {
			a, _ := s.Pop()
			b, err := s.Pop()
			if err != nil {
				return err
			}
			if a == 0 {
				return ErrDivideByZero
			}
			s.Push(b / a)
			return nil
		},
		"dup": func(s *stack) error {
			top, err := s.Top()
			if err != nil {
				return err
			}
			s.Push(top)
			return nil
		},
		"over": func(s *stack) error {
			_, err := s.Over()
			return err
		},
		"drop": func(s *stack) error {
			_, err := s.Pop()
			return err
		},
		"swap": func(s *stack) error {
			return s.Swap()
		},
	}
	NewEvaluator.KnownOperation = knownOperation
	return &NewEvaluator
}

func (e *Evaluator) ProcessValue(value int) {
	e.Stack.Push(value)
}

func (e *Evaluator) tokenToOperation(token string) (func(*stack) error, error) {
	if op, ok := e.KnownOperation[token]; ok {
		return op, nil
	}
	if n, err := strconv.Atoi(token); err == nil {
		return pushConstant(n), nil
	}
	return nil, ErrUnknownToken(token)
}

func (e *Evaluator) parseFunctionBody() ([]func(*stack) error, error) {
	var ops []func(*stack) error
	for {
		word := e.ReadWord()
		if word == "" {
			return nil, ErrInvalidFunctionDefinition
		}
		if word == ";" {
			break
		}
		if word == ":" {
			if err := e.AddFunction(); err != nil {
				return nil, err
			}
			continue
		}
		op, err := e.tokenToOperation(word)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func (e *Evaluator) AddFunction() error {
	fname := e.ReadWord()

	if fname == "" || isNumeric(fname) {
		return ErrInvalidFunctionDefinition
	}

	operationsSequence, err := e.parseFunctionBody()
	if err != nil {
		return err
	}

	e.KnownOperation[fname] = func(s *stack) error {
		for _, op := range operationsSequence {
			if err := op(s); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func (e *Evaluator) AddRow(row string) error {
	e.Row = append(e.Row, strings.Fields(row)...)
	return nil
}

func (e *Evaluator) ProcessOperation(fname string) error {
	f := e.KnownOperation[fname]
	return f(e.Stack)
}

func (e *Evaluator) ProcessWord(word string) error {
	if word == ":" {
		return e.AddFunction()
	} else if _, ok := e.KnownOperation[word]; ok {
		return e.ProcessOperation(word)
	} else if intValue, err := strconv.Atoi(word); err == nil {
		e.ProcessValue(intValue)
		return err
	}
	return ErrUnprocessableEntity(word)
}

func (e *Evaluator) Process(row string) ([]int, error) {

	e.Row = strings.Fields(row)
	for word := e.ReadWord(); word != ""; word = e.ReadWord() {
		err := e.ProcessWord(word)
		if err != nil {
			return []int{}, err
		}
	}
	return *e.Stack, nil
}
