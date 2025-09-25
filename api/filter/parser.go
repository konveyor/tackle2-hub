package filter

import (
	"math"
	"strconv"
	"strings"
)

// Parser used to parse the filter.
type Parser struct {
}

// Filter parses the filter and builds a Filter.
func (r *Parser) Filter(filter string) (f Filter, err error) {
	if filter == "" {
		return
	}
	lexer := Lexer{}
	err = lexer.With(string(COMMA) + filter)
	if err != nil {
		return
	}
	var bfr []Token
	for {
		token, next := lexer.next()
		if !next {
			break
		}
		if len(bfr) > 2 {
			if bfr[0].Kind != OPERATOR || bfr[2].Kind != OPERATOR {
				err = Errorf("Syntax error.")
				return
			}
			switch token.Kind {
			case LITERAL, STRING:
				p := Predicate{
					Unused:   bfr[0],
					Field:    bfr[1],
					Operator: bfr[2],
					Value:    Value{token},
				}
				f.predicates = append(f.predicates, p)
				bfr = nil
			case LPAREN:
				lexer.put()
				list := List{&lexer}
				v, nErr := list.Build()
				if nErr != nil {
					err = nErr
					return
				}
				p := Predicate{
					Unused:   bfr[0],
					Field:    bfr[1],
					Operator: bfr[2],
					Value:    v,
				}
				f.predicates = append(f.predicates, p)
				bfr = nil
			}
		} else {
			bfr = append(bfr, token)
		}
	}
	if len(bfr) != 0 {
		err = Errorf("Syntax error.")
		return
	}
	return
}

// Predicate filter predicate.
type Predicate struct {
	Unused   Token
	Field    Token
	Operator Token
	Value    Value
}

// Value term value.
type Value []Token

// ByKind returns values by kind.
func (r Value) ByKind(kind ...byte) (matched []Token) {
	for _, t := range r {
		for _, k := range kind {
			if t.Kind == k {
				matched = append(matched, t)
			}
		}
	}
	return
}

// Operator returns true when contains the specified operator.
func (r *Value) Operator(operator byte) (matched bool) {
	operators := r.ByKind(OPERATOR)
	if len(operators) > 0 {
		matched = operators[0].Value[0] == operator
	}
	return
}

// Join values with operator.
func (r *Value) Join(operator byte) (out Value) {
	for i := range r.ByKind(LITERAL, STRING) {
		if i > 0 {
			out = append(out, Token{Kind: OPERATOR, Value: string(operator)})
		}
		out = append(out, (*r)[i])
	}
	return
}

func (r *Value) As(x any) {
	switch x := x.(type) {
	case *[]string:
		for _, t := range *r {
			if t.Kind != OPERATOR {
				*x = append(*x, t.Value)
			}
		}
	case *[]int:
		for _, t := range *r {
			if t.Kind != OPERATOR {
				n, _ := strconv.Atoi(t.Value)
				*x = append(*x, n)
			}
		}
	}
	return
}

// Pj returns a postgres json value renderer.
func (r *Value) Pj() (pj *Pj) {
	pj = &Pj{Value: r}
	return
}

// Pj is a postgres json value renderer.
type Pj struct {
	*Value
}

// Array render a postgres array[].
func (p *Pj) Array(x any) (j string) {
	p.As(x)
	sv := make([]string, 0)
	switch x := x.(type) {
	case *[]string:
		for i := range *x {
			s := "'"
			s += (*x)[i]
			s += "'"
			sv = append(sv, s)
		}
	case *[]int:
		for _, n := range *x {
			sv = append(sv, strconv.Itoa(n))
		}
	}
	j = "array["
	j += strings.Join(sv, ",")
	j += "]"
	return
}

// LitArray renders a (literal) postgres json array.
func (p *Pj) LitArray(x any) (j string) {
	p.As(x)
	sv := make([]string, 0)
	switch x := x.(type) {
	case *[]string:
		for i := range *x {
			s := "\""
			s += (*x)[i]
			s += "\""
			sv = append(sv, s)
		}
	case *[]int:
		for _, n := range *x {
			sv = append(sv, strconv.Itoa(n))
		}
	}
	j = "'["
	j += strings.Join(sv, ",")
	j += "]'"
	return
}

// List construct.
// Example: (red|blue|green)
type List struct {
	*Lexer
}

// Build the value.
func (r *List) Build() (v Value, err error) {
	for {
		token, next := r.next()
		if !next {
			err = Errorf("End ')' not found.")
			break
		}
		switch token.Kind {
		case LITERAL, STRING:
			v = append(v, token)
		case OPERATOR:
			switch token.Value {
			case string(AND),
				string(OR):
				v = append(v, token)
			default:
				err = Errorf("List separator must be `,` `|`")
				return
			}
		case LPAREN:
			// ignored.
		case RPAREN:
			err = r.validate(v)
			return
		default:
			err = Errorf("'%s' not expected in ()", token.Value)
			return
		}
	}

	return
}

// validate the result.
func (r *List) validate(v Value) (err error) {
	lastOp := byte(0)
	for i := range v {
		if math.Mod(float64(i), 2) == 0 {
			switch v[i].Kind {
			case LITERAL,
				STRING:
			default:
				err = Errorf("(LITERAL|STRING) not expected in ()")
				return
			}
		} else {
			switch v[i].Kind {
			case OPERATOR:
				operator := v[i].Value[0]
				if lastOp != 0 {
					if operator != lastOp {
						err = Errorf("Mixed operator detected in ().")
						return
					}
				}
				lastOp = operator
			default:
				err = Errorf("OPERATOR expected in ()")
				return
			}
		}
	}
	if len(v) == 0 {
		err = Errorf("List cannot be empty.")
	}
	return
}
