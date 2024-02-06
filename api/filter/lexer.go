package filter

const (
	EOF    = 0
	COLON  = ':'
	COMMA  = ','
	AND    = COMMA
	OR     = '|'
	EQ     = '='
	LIKE   = '~'
	NOT    = '!'
	LT     = '<'
	GT     = '>'
	QUOTE  = '"'
	SQUOTE = '\''
	ESCAPE = '\\'
	SPACE  = ' '
	LPAREN = '('
	RPAREN = ')'
)

const (
	LITERAL  = 0x01
	STRING   = 0x02
	OPERATOR = 0x03
)

// Lexer token reader.
type Lexer struct {
	tokens []Token
	index  int
}

// With builds with the specified filter.
func (r *Lexer) With(filter string) (err error) {
	reader := Reader{input: filter}
	var bfr []byte
	push := func(kind byte) {
		if len(bfr) == 0 {
			return
		}
		r.tokens = append(
			r.tokens,
			Token{
				Kind:  kind,
				Value: string(bfr),
			})
		bfr = nil
	}
	for {
		ch := reader.next()
		if ch == EOF {
			break
		}
		switch ch {
		case QUOTE,
			SQUOTE:
			reader.put()
			push(LITERAL)
			q := Quoted{Reader: &reader}
			token, nErr := q.Read()
			if nErr != nil {
				err = nErr
				return
			}
			r.tokens = append(r.tokens, token)
		case SPACE:
			push(LITERAL)
		case LPAREN:
			bfr = append(bfr, ch)
			push(LPAREN)
		case RPAREN:
			push(LITERAL)
			bfr = append(bfr, ch)
			push(RPAREN)
		case COLON,
			COMMA,
			OR,
			EQ,
			LIKE,
			NOT,
			LT,
			GT:
			reader.put()
			push(LITERAL)
			q := Operator{Reader: &reader}
			token, nErr := q.Read()
			if nErr != nil {
				err = nErr
				return
			}
			r.tokens = append(r.tokens, token)
		default:
			bfr = append(bfr, ch)
		}
	}
	push(LITERAL)
	return
}

// next returns the next token.
func (r *Lexer) next() (token Token, next bool) {
	if r.index < len(r.tokens) {
		token = r.tokens[r.index]
		next = true
		r.index++
	}
	return
}

// Put rewinds the lexer by 1 token.
func (r *Lexer) put() {
	if r.index > 0 {
		r.index--
	}
}

// Token scanned token.
type Token struct {
	Kind  byte
	Value string
}

// Reader scan the input.
type Reader struct {
	input string
	index int
}

// Next character.
// Returns 0 at EOF.
func (r *Reader) next() (ch byte) {
	if r.index < len(r.input) {
		ch = r.input[r.index]
		r.index++
	}
	return
}

// Put rewinds one character.
func (r *Reader) put() {
	if r.index > 0 {
		r.index--
	}
}

// Quoted string token reader.
type Quoted struct {
	*Reader
}

// Read token.
func (q *Quoted) Read() (token Token, err error) {
	lastCh := byte(0)
	var bfr []byte
	quote := q.next()
	for {
		ch := q.next()
		if ch == EOF {
			break
		}
		switch ch {
		case quote:
			if lastCh != ESCAPE {
				token.Kind = STRING
				token.Value = string(bfr)
				return
			}
		default:
			bfr = append(bfr, ch)
		}
		lastCh = ch
	}
	err = Errorf("End (%c) not found.", quote)
	return
}

// Operator token reader.
type Operator struct {
	*Reader
}

// Read token.
func (q *Operator) Read() (token Token, err error) {
	var bfr []byte
	for {
		ch := q.next()
		if ch == EOF {
			break
		}
		switch ch {
		case COLON,
			COMMA,
			OR,
			EQ,
			LIKE,
			NOT,
			LT,
			GT:
			bfr = append(bfr, ch)
		default:
			q.put()
			token.Kind = OPERATOR
			token.Value = string(bfr)
			return
		}
	}
	err = Errorf("End of operator not found.")
	return
}
