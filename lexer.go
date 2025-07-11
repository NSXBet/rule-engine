package rule

import (
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	input    string
	runes    []rune
	position int
	current  rune
	tokens   []Token
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		runes:  []rune(input),
		tokens: make([]Token, 0, tokenSliceInitialCapacity),
	}
	l.readChar()

	return l
}

func (l *Lexer) Tokenize() []Token {
	for l.current != 0 {
		l.skipWhitespace()

		if l.current == 0 {
			break
		}

		start := l.position - 1

		switch l.current {
		case '(':
			l.handleSingleCharToken(PAREN_OPEN, start)
		case ')':
			l.handleSingleCharToken(PAREN_CLOSE, start)
		case '[':
			l.handleSingleCharToken(ARRAY_START, start)
		case ']':
			l.handleSingleCharToken(ARRAY_END, start)
		case '.':
			l.handleSingleCharToken(DOT, start)
		case ',':
			l.handleSingleCharToken(COMMA, start)
		case '"':
			l.handleStringToken(start)
		case '=':
			l.handleEqualsToken(start)
		case '!':
			l.handleNotEqualsToken(start)
		case '-':
			l.handleMinusToken(start)
		default:
			l.handleDefaultToken(start)
		}
	}

	l.tokens = append(l.tokens, Token{Type: EOF, Start: l.position, End: l.position})

	return l.tokens
}

func (l *Lexer) handleSingleCharToken(tokenType TokenType, start int) {
	l.tokens = append(l.tokens, Token{Type: tokenType, Start: start, End: l.position})
	l.readChar()
}

func (l *Lexer) handleStringToken(start int) {
	value := l.readString()
	l.tokens = append(l.tokens, Token{
		Type:  STRING,
		Value: value,
		Start: start,
		End:   l.position - 1,
	})
}

func (l *Lexer) handleEqualsToken(start int) {
	if l.peekChar() == '=' {
		l.readChar()
		l.readChar()
		l.tokens = append(l.tokens, Token{Type: EQUALS, Start: start, End: l.position - 1})
	} else {
		l.readChar()
	}
}

func (l *Lexer) handleNotEqualsToken(start int) {
	if l.peekChar() == '=' {
		l.readChar()
		l.readChar()
		l.tokens = append(l.tokens, Token{Type: NOT_EQUALS, Start: start, End: l.position - 1})
	} else {
		l.readChar()
	}
}

func (l *Lexer) handleMinusToken(start int) {
	// Check if this is a negative number
	if unicode.IsDigit(l.peekChar()) {
		l.readChar() // consume the '-'
		value, num, isLargeInt := l.readNumber()
		// Make it negative
		value = "-" + value
		num = -num

		if isLargeInt {
			// Store large integers as strings to preserve precision
			l.tokens = append(l.tokens, Token{
				Type:  STRING,
				Value: value,
				Start: start,
				End:   l.position - 1,
			})
		} else {
			l.tokens = append(l.tokens, Token{
				Type:     NUMBER,
				Value:    value,
				NumValue: num,
				Start:    start,
				End:      l.position - 1,
			})
		}
	} else {
		// Just a minus operator
		l.readChar()
	}
}

func (l *Lexer) handleDefaultToken(start int) {
	switch {
	case unicode.IsDigit(l.current):
		l.handleNumberToken(start)
	case unicode.IsLetter(l.current) || l.current == '_':
		l.handleIdentifierToken(start)
	default:
		l.readChar()
	}
}

func (l *Lexer) handleNumberToken(start int) {
	value, num, isLargeInt := l.readNumber()
	if isLargeInt {
		// Store large integers as strings to preserve precision
		l.tokens = append(l.tokens, Token{
			Type:  STRING,
			Value: value,
			Start: start,
			End:   l.position - 1,
		})
	} else {
		l.tokens = append(l.tokens, Token{
			Type:     NUMBER,
			Value:    value,
			NumValue: num,
			Start:    start,
			End:      l.position - 1,
		})
	}
}

func (l *Lexer) handleIdentifierToken(start int) {
	value := l.readIdentifier()
	tokenType := IDENTIFIER

	if kwType, exists := keywordMap[value]; exists {
		tokenType = kwType
		if tokenType == BOOLEAN {
			boolVal := value == trueString
			l.tokens = append(l.tokens, Token{
				Type:      BOOLEAN,
				Value:     value,
				BoolValue: boolVal,
				Start:     start,
				End:       l.position - 1,
			})

			return
		}
	}

	l.tokens = append(l.tokens, Token{
		Type:  tokenType,
		Value: value,
		Start: start,
		End:   l.position - 1,
	})
}

func (l *Lexer) readChar() {
	if l.position >= len(l.runes) {
		l.current = 0
	} else {
		l.current = l.runes[l.position]
	}

	l.position++
}

func (l *Lexer) peekChar() rune {
	if l.position >= len(l.runes) {
		return 0
	}

	return l.runes[l.position]
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.current) {
		l.readChar()
	}
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening quote

	var result strings.Builder

	for l.current != '"' && l.current != 0 {
		if l.current == '\\' {
			// Handle escape sequences
			l.readChar() // consume backslash

			if l.current == 0 {
				break // End of input
			}

			switch l.current {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			default:
				// For unrecognized escape sequences, include the backslash and character
				result.WriteByte('\\')
				result.WriteRune(l.current)
			}
		} else {
			result.WriteRune(l.current)
		}

		l.readChar()
	}

	l.readChar() // skip closing quote

	return result.String()
}

func (l *Lexer) readNumber() (string, float64, bool) {
	start := l.position - 1

	for unicode.IsDigit(l.current) || l.current == '.' {
		l.readChar()
	}

	str := string(l.runes[start : l.position-1])
	num, _ := strconv.ParseFloat(str, 64)

	// Check if this is a large integer that would lose precision
	isLargeInt := l.isLargeInteger(str)

	return str, num, isLargeInt
}

func (l *Lexer) isLargeInteger(s string) bool {
	// Check if it's an integer (no decimal point)
	hasDecimal := false

	for _, r := range s {
		if r == '.' {
			hasDecimal = true
			break
		}
	}

	if hasDecimal {
		return false
	}

	// Parse as int64 to check if it's a large integer
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Check if it would lose precision when converted to float64
		return val > maxSafeInteger || val < minSafeInteger
	}

	return false
}

func (l *Lexer) readIdentifier() string {
	start := l.position - 1

	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		l.readChar()
	}

	return string(l.runes[start : l.position-1])
}
