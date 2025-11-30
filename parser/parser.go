package parser

import (
	"regexp"
	"strings"
	"fmt"
)

type Token struct {
	Value string
	Depth int
}

func NewToken(value string, depth int) *Token {
	return &Token {
		Value: value,
		Depth: depth,
	}
}

func ParseJSON(json string) (bool, error) {
	isObjectOrArray := false

	objectRe := regexp.MustCompile(`(?s)^{(.*)}$`)
	arrayRe := regexp.MustCompile(`(?s)^\[(.*)\]$`)
	stringRe := regexp.MustCompile(`(?s)^\"(?:[^\\\n\t]*(?:\\(?:\"|\\|\/|b|f|n|r|t|u|[1-9]))*)*\"$`)
	numberRe := regexp.MustCompile(`^[\+-]?(?:\d|[1-9]\d+)(?:\.\d+)?(?:[Ee][\+-]?\d+)?$`)
	boolRe := regexp.MustCompile(`^(?:true|false)$`)
	nullRe := regexp.MustCompile(`^null$`)

	jsonStack := NewStack()
	jsonStack.Push(NewToken(json, -1))
	for {
		if jsonStack.IsEmpty() {
			break
		}

		token := jsonStack.Pop().(*Token)
		
		if token.Depth > 18 {
			return false, fmt.Errorf("too deep")
		}

		token.Value = strings.TrimSpace(token.Value)
		if token.Value == "" {
			continue
		}

		objectMatch := objectRe.FindStringSubmatch(token.Value)
		if objectMatch != nil {
			token.Depth++
			isObjectOrArray = true

			isValidObject, err := parseObject(jsonStack, []rune(objectMatch[1]), token.Depth)
			if !isValidObject || err != nil {
				return false, err
			}

			continue
		}

		arrayMatch := arrayRe.FindStringSubmatch(token.Value)
		if arrayMatch != nil {
			token.Depth++
			isObjectOrArray = true

			isValidArray, err := parseArray(jsonStack, []rune(arrayMatch[1]), token.Depth)
			if !isValidArray || err != nil {
				return false, err
			}

			continue
		}

		if stringRe.MatchString(token.Value) {
			continue
		}		

		if numberRe.MatchString(token.Value) {
			continue
		}

		if boolRe.MatchString(token.Value) {
			continue
		}

		if nullRe.MatchString(token.Value) {
			continue
		}

		return false, fmt.Errorf("invalid json %s", token.Value)
	}

	if !isObjectOrArray {
		return false, fmt.Errorf("json is neither object nor array")
	}

	return true, nil
}

func parseObject(jsonStack *Stack, json []rune, depth int) (bool, error) {	
	objectStack := NewStack()
	var sb strings.Builder
	escapeFlag := false
	commaFlag := false
	for i := 0; i < len(json); i++ {
		if json[i] != ',' || !objectStack.IsEmpty() {
			sb.WriteRune(json[i])
		}

		if escapeFlag {
			escapeFlag = false
			continue
		}

		switch json[i] {
			case '{', '[', '"':
				if objectStack.Peek() != '"' {
					objectStack.Push(json[i])
				} else if json[i] == '"' && objectStack.Peek() == '"' {
					objectStack.Pop()
				}
			case '}':
				if objectStack.Peek() == '{' {
					objectStack.Pop()
				}
			case ']':
				if objectStack.Peek() == '[' {
					objectStack.Pop()
				}
			case '\\':
				escapeFlag = true
			case ',':
				if objectStack.IsEmpty() {
					commaFlag = true
					
					isValidKVP, err := parseKVP(jsonStack, sb.String(), depth, commaFlag)
					if !isValidKVP || err != nil {
						return false, err
					} 
					sb.Reset()					
				}
		}	
	}

	if sb.Len() > 0 {
		isValidKVP, err := parseKVP(jsonStack, sb.String(), depth, commaFlag)
		if !isValidKVP || err != nil {
			return false, err
		}
	}

	if len(json) > 0 && json[len(json)-1] == ',' {
		return false, fmt.Errorf("extra comma")
	}

	return true, nil
}

func parseKVP(jsonStack *Stack, json string, depth int, commaFlag bool) (bool, error) {
	json = strings.TrimSpace(json)
	if json == "" {
		if commaFlag {
			return false, fmt.Errorf("extra comma")
		}

		return true, nil
	}

	kvpRe := regexp.MustCompile(`(?s)^\".*?(?:\"\s*:)(.*)$`)
	kvpMatch := kvpRe.FindStringSubmatch(json)
	if kvpMatch != nil {
		jsonStack.Push(NewToken(kvpMatch[1], depth))
		return true, nil
	}

	return false, fmt.Errorf("invalid kvp %s", json)
}

func parseArray(jsonStack *Stack, json []rune, depth int) (bool, error) {	
	arrayStack := NewStack()
	var sb strings.Builder
	escapeFlag := false
	commaFlag := false
	for i := 0; i < len(json); i++ {
		if json[i] != ',' || !arrayStack.IsEmpty() {
			sb.WriteRune(json[i])
		}
		
		if escapeFlag {
			escapeFlag = false
			continue
		}

		switch json[i] {
			case '{', '[', '"':
				if arrayStack.Peek() != '"' {
					arrayStack.Push(json[i])
				} else if json[i] == '"' && arrayStack.Peek() == '"' {
					arrayStack.Pop()
				}
			case '}':
				if arrayStack.Peek() == '{' {
					arrayStack.Pop()
				}
			case ']':
				if arrayStack.Peek() == '[' {
					arrayStack.Pop()
				}
			case '\\':
				escapeFlag = true
			case ',':
				if arrayStack.IsEmpty() {
					commaFlag = true
					
					value := strings.TrimSpace(sb.String())
					if value == "" {
						return false, fmt.Errorf("extra comma")
					}

					jsonStack.Push(NewToken(value, depth))	 
					sb.Reset()
				}
		}	
	}

	if sb.Len() > 0 {
		value := strings.TrimSpace(sb.String())
		if value == "" && commaFlag {
			return false, fmt.Errorf("extra comma")
		} else if value != "" {
			jsonStack.Push(NewToken(value, depth))
		}
	}

	if len(json) > 0 && json[len(json)-1] == ',' {
		return false, fmt.Errorf("extra comma")
	}
		
	return true, nil
}
