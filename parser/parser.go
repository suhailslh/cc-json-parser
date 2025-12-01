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

func IsValidJSON(json string) (bool, error) {
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
		if token.Value == "" && token.Depth > -1 {
			continue
		}

		objectMatch := objectRe.FindStringSubmatch(token.Value)
		if objectMatch != nil {
			isObjectOrArray = true
			token.Depth++

			kvpTokens, err := parseObject(NewToken(objectMatch[1], token.Depth))
			if err != nil {
				return false, err
			}

			for _, kvpToken := range kvpTokens {
				jsonStack.Push(kvpToken)
			}

			continue
		}

		arrayMatch := arrayRe.FindStringSubmatch(token.Value)
		if arrayMatch != nil {
			isObjectOrArray = true
			token.Depth++

			valueTokens, err := parseArray(NewToken(arrayMatch[1], token.Depth))
			if err != nil {
				return false, err
			}

			for _, valueToken := range valueTokens {
				jsonStack.Push(valueToken)
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

func parseObject(token *Token) ([]*Token, error) {	
	var result []*Token
	json := []rune(token.Value)
	objectStack := NewStack()
	var sb strings.Builder
	escapeFlag := false
	commaFlag := false

	for _, char := range json {
		if char != ',' || !objectStack.IsEmpty() {
			sb.WriteRune(char)
		}

		if escapeFlag {
			escapeFlag = false
			continue
		}

		switch char {
			case '{', '[':
				if objectStack.Peek() != '"' {
					objectStack.Push(char)
				}
			case '"':
				if objectStack.Peek() != '"' {
					objectStack.Push(char)
				} else {
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
					
					kvpTokens, err := parseKVP(NewToken(sb.String(), token.Depth), commaFlag)
					if err != nil {
						return result, err
					} 
					
					for _, kvpToken := range kvpTokens {
						result = append(result, kvpToken)
					}

					sb.Reset()					
				}
		}	
	}

	if sb.Len() > 0 {
		kvpTokens, err := parseKVP(NewToken(sb.String(), token.Depth), commaFlag)
		if err != nil {
			return result, err
		}

		for _, kvpToken := range kvpTokens {
			result = append(result, kvpToken)
		}
	}

	if len(json) > 0 && json[len(json)-1] == ',' {
		return result, fmt.Errorf("extra comma")
	}

	return result, nil
}

func parseKVP(token *Token, commaFlag bool) ([]*Token, error) {
	var result []*Token
	json := strings.TrimSpace(token.Value)
	
	if json == "" {
		if commaFlag {
			return result, fmt.Errorf("extra comma")
		}

		return result, nil
	}

	kvpRe := regexp.MustCompile(`(?s)^\".*?(?:\"\s*:)(.*)$`)
	kvpMatch := kvpRe.FindStringSubmatch(json)
	if kvpMatch != nil {
		result = append(result, NewToken(kvpMatch[1], token.Depth))
		return result, nil
	}

	return result, fmt.Errorf("invalid kvp %s", json)
}

func parseArray(token *Token) ([]*Token, error) {	
	var result []*Token
	json := []rune(token.Value)
	arrayStack := NewStack()
	var sb strings.Builder
	escapeFlag := false

	for _, char := range json {
		if char != ',' || !arrayStack.IsEmpty() {
			sb.WriteRune(char)
		}
		
		if escapeFlag {
			escapeFlag = false
			continue
		}

		switch char {
			case '{', '[':
				if arrayStack.Peek() != '"' {
					arrayStack.Push(char)
				}
			case '"':
				if arrayStack.Peek() != '"' {
					arrayStack.Push(char)
				} else {
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
					value := strings.TrimSpace(sb.String())
					if value == "" {
						return result, fmt.Errorf("extra comma")
					}

					result = append(result, NewToken(value, token.Depth))
					sb.Reset()
				}
		}	
	}

	if sb.Len() > 0 {
		result = append(result, NewToken(sb.String(), token.Depth))
	}

	if len(json) > 0 && json[len(json)-1] == ',' {
		return result, fmt.Errorf("extra comma")
	}
		
	return result, nil
}
