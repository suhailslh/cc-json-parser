package parser

import (
	"regexp"
	"strings"
	"fmt"
)

const maxDepth = 18

type token struct {
	Value string
	Depth int
}

func newToken(value string, depth int) *token {
	return &token {
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
	jsonStack.Push(newToken(json, -1))
	for {
		if jsonStack.IsEmpty() {
			break
		}

		tk := jsonStack.Pop().(*token)
		
		tk.Value = strings.TrimSpace(tk.Value)

		objectMatch := objectRe.FindStringSubmatch(tk.Value)
		if objectMatch != nil {
			isObjectOrArray = true
			tk.Depth++

			if tk.Depth > maxDepth {
				return false, fmt.Errorf("too deep")
			}

			kvpTokens, err := parseObject(newToken(objectMatch[1], tk.Depth))
			if err != nil {
				return false, err
			}

			for _, kvpToken := range kvpTokens {
				jsonStack.Push(kvpToken)
			}

			continue
		}

		arrayMatch := arrayRe.FindStringSubmatch(tk.Value)
		if arrayMatch != nil {
			isObjectOrArray = true
			tk.Depth++

			if tk.Depth > maxDepth {
				return false, fmt.Errorf("too deep")
			}

			valueTokens, err := parseArray(newToken(arrayMatch[1], tk.Depth))
			if err != nil {
				return false, err
			}

			for _, valueToken := range valueTokens {
				jsonStack.Push(valueToken)
			}

			continue
		}

		if stringRe.MatchString(tk.Value) {
			continue
		}		

		if numberRe.MatchString(tk.Value) {
			continue
		}

		if boolRe.MatchString(tk.Value) {
			continue
		}

		if nullRe.MatchString(tk.Value) {
			continue
		}

		return false, fmt.Errorf("invalid json %s", tk.Value)
	}

	if !isObjectOrArray {
		return false, fmt.Errorf("json is neither object nor array")
	}

	return true, nil
}

func parseObject(tk *token) ([]*token, error) {	
	var result []*token
	json := []rune(strings.TrimSpace(tk.Value))
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
					
					kvpTokens, err := parseKVP(newToken(sb.String(), tk.Depth), commaFlag)
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
		kvpTokens, err := parseKVP(newToken(sb.String(), tk.Depth), commaFlag)
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

func parseKVP(tk *token, commaFlag bool) ([]*token, error) {
	var result []*token
	json := strings.TrimSpace(tk.Value)
	
	if json == "" {
		if commaFlag {
			return result, fmt.Errorf("extra comma")
		}

		return result, nil
	}

	kvpRe := regexp.MustCompile(`(?s)^\".*?(?:\"\s*:)(.*)$`)
	kvpMatch := kvpRe.FindStringSubmatch(json)
	if kvpMatch != nil {
		result = append(result, newToken(kvpMatch[1], tk.Depth))
		return result, nil
	}

	return result, fmt.Errorf("invalid kvp %s", json)
}

func parseArray(tk *token) ([]*token, error) {	
	var result []*token
	json := []rune(strings.TrimSpace(tk.Value))
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

					result = append(result, newToken(value, tk.Depth))
					sb.Reset()
				}
		}	
	}

	if sb.Len() > 0 {
		result = append(result, newToken(sb.String(), tk.Depth))
	}

	if len(json) > 0 && json[len(json)-1] == ',' {
		return result, fmt.Errorf("extra comma")
	}
		
	return result, nil
}
