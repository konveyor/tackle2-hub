package filter

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestLexer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	lexer := Lexer{}
	err := lexer.With("name:elmer,age:20")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: LITERAL, Value: "elmer"},
			{Kind: OPERATOR, Value: string(COMMA)},
			{Kind: LITERAL, Value: "age"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: LITERAL, Value: "20"},
		}))

	lexer = Lexer{}
	err = lexer.With("name:\"one|two\"")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: STRING, Value: "one|two"},
		}))

	lexer = Lexer{}
	err = lexer.With("name:\"one=two\"")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: STRING, Value: "one=two"},
		}))

	lexer = Lexer{}
	err = lexer.With("name:\"(one|two)\"")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: STRING, Value: "(one|two)"},
		}))

	lexer = Lexer{}
	err = lexer.With("name:\"hello world\"")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: STRING, Value: "hello world"},
		}))

	lexer = Lexer{}
	err = lexer.With("name = \"elmer\" , age > 20")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(EQ)},
			{Kind: STRING, Value: "elmer"},
			{Kind: OPERATOR, Value: string(COMMA)},
			{Kind: LITERAL, Value: "age"},
			{Kind: OPERATOR, Value: string(GT)},
			{Kind: LITERAL, Value: "20"},
		}))

	lexer = Lexer{}
	err = lexer.With("name~elmer*")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(LIKE)},
			{Kind: LITERAL, Value: "elmer*"},
		}))

	lexer = Lexer{}
	err = lexer.With("name=(one|two|three)")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(EQ)},
			{Kind: LPAREN, Value: string(LPAREN)},
			{Kind: LITERAL, Value: "one"},
			{Kind: OPERATOR, Value: string(OR)},
			{Kind: LITERAL, Value: "two"},
			{Kind: OPERATOR, Value: string(OR)},
			{Kind: LITERAL, Value: "three"},
			{Kind: RPAREN, Value: string(RPAREN)},
		}))

	lexer = Lexer{}
	err = lexer.With("name=(one,two,three)")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(EQ)},
			{Kind: LPAREN, Value: string(LPAREN)},
			{Kind: LITERAL, Value: "one"},
			{Kind: OPERATOR, Value: string(COMMA)},
			{Kind: LITERAL, Value: "two"},
			{Kind: OPERATOR, Value: string(COMMA)},
			{Kind: LITERAL, Value: "three"},
			{Kind: RPAREN, Value: string(RPAREN)},
		}))

	lexer = Lexer{}
	err = lexer.With("name:'elmer'")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(lexer.tokens).To(gomega.Equal(
		[]Token{
			{Kind: LITERAL, Value: "name"},
			{Kind: OPERATOR, Value: string(COLON)},
			{Kind: STRING, Value: "elmer"},
		}))
}

func TestParser(t *testing.T) {
	var err error
	g := gomega.NewGomegaWithT(t)

	p := Parser{}
	_, err = p.Filter("")
	g.Expect(err).To(gomega.BeNil())

	p = Parser{}
	f, err := p.Filter("name:elmer,age:20")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f.predicates).To(gomega.Equal(
		[]Predicate{
			{
				Unused:   Token{Kind: OPERATOR, Value: string(COMMA)},
				Field:    Token{Kind: LITERAL, Value: "name"},
				Operator: Token{Kind: OPERATOR, Value: string(COLON)},
				Value:    Value{Token{Kind: LITERAL, Value: "elmer"}},
			},
			{
				Unused:   Token{Kind: OPERATOR, Value: string(COMMA)},
				Field:    Token{Kind: LITERAL, Value: "age"},
				Operator: Token{Kind: OPERATOR, Value: string(COLON)},
				Value:    Value{Token{Kind: LITERAL, Value: "20"}},
			},
		}))

	p = Parser{}
	f, err = p.Filter("name:elmer,category=(one|two|three),age:20")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f.predicates).To(gomega.Equal(
		[]Predicate{
			{
				Unused:   Token{Kind: OPERATOR, Value: string(COMMA)},
				Field:    Token{Kind: LITERAL, Value: "name"},
				Operator: Token{Kind: OPERATOR, Value: string(COLON)},
				Value:    Value{Token{Kind: LITERAL, Value: "elmer"}},
			},
			{
				Unused:   Token{Kind: OPERATOR, Value: string(COMMA)},
				Field:    Token{Kind: LITERAL, Value: "category"},
				Operator: Token{Kind: OPERATOR, Value: string(EQ)},
				Value: Value{
					Token{Kind: LITERAL, Value: "one"},
					Token{Kind: OPERATOR, Value: string(OR)},
					Token{Kind: LITERAL, Value: "two"},
					Token{Kind: OPERATOR, Value: string(OR)},
					Token{Kind: LITERAL, Value: "three"},
				},
			},
			{
				Unused:   Token{Kind: OPERATOR, Value: string(COMMA)},
				Field:    Token{Kind: LITERAL, Value: "age"},
				Operator: Token{Kind: OPERATOR, Value: string(COLON)},
				Value:    Value{Token{Kind: LITERAL, Value: "20"}},
			},
		}))

	p = Parser{}
	f, err = p.Filter("cat=()")
	g.Expect(err).ToNot(gomega.BeNil())

	p = Parser{}
	f, err = p.Filter("cat=(one|two,three)")
	g.Expect(err).ToNot(gomega.BeNil())
}

func TestFilter(t *testing.T) {
	var err error
	g := gomega.NewGomegaWithT(t)
	p := Parser{}

	filter, err := p.Filter("")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(filter.predicates)).To(gomega.Equal(0))
	g.Expect(filter.Empty()).To(gomega.BeTrue())

	filter, err = p.Filter("name:elmer,age:20,category=(a|b|c),name.first:elmer,name.last=fudd")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(filter.predicates)).To(gomega.Equal(5))
	g.Expect(filter.Empty()).To(gomega.BeFalse())

	f, found := filter.Field("name")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(f.Name()).To(gomega.Equal("name"))
	g.Expect(AsValue(f.Value[0])).To(gomega.Equal("elmer"))

	f, found = filter.Field("category")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(f.Name()).To(gomega.Equal("category"))
	g.Expect(f.Value.ByKind(OPERATOR)).To(gomega.Equal([]Token{
		{Kind: OPERATOR, Value: string(OR)},
		{Kind: OPERATOR, Value: string(OR)},
	}))
	g.Expect(f.Value.ByKind(LITERAL, STRING)).To(gomega.Equal([]Token{
		{Kind: LITERAL, Value: "a"},
		{Kind: LITERAL, Value: "b"},
		{Kind: LITERAL, Value: "c"},
	}))
	sql, values := f.SQL()
	g.Expect(sql).To(gomega.Equal("category IN ?"))
	g.Expect(values[0]).To(gomega.Equal([]any{"a", "b", "c"}))

	f, found = filter.Field("name.first")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(AsValue(f.Value[0])).To(gomega.Equal("elmer"))
	f, found = filter.Field("name.last")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(AsValue(f.Value[0])).To(gomega.Equal("fudd"))

	r := filter.Resource("name")
	f, found = r.Field("first")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(f.Name()).To(gomega.Equal("first"))
	g.Expect(AsValue(f.Value[0])).To(gomega.Equal("elmer"))

	r = filter.Resource("Name")
	f, found = r.Field("First")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(f.Name()).To(gomega.Equal("first"))

	filter, err = p.Filter("app.name=test,app.tag.id=0")
	filter = filter.Resource("app")
	f, found = filter.Field("name")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(AsValue(f.Value[0])).To(gomega.Equal("test"))
	filter = filter.Resource("tag")
	f, found = filter.Field("id")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(AsValue(f.Value[0])).To(gomega.Equal(0))

	filter, err = p.Filter("category~(a|b)")
	f, found = filter.Field("category")
	g.Expect(found).To(gomega.BeTrue())
	sql, values = f.SQL()
	g.Expect(sql).To(gomega.Equal("(category LIKE ? OR category LIKE ?)"))
	g.Expect(values).To(gomega.Equal([]any{"a", "b"}))
}

func TestValidation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	p := Parser{}
	filter, err := p.Filter("name:elmer,age>20,category=(a|b|c)")
	g.Expect(err).To(gomega.BeNil())
	err = filter.Validate(
		[]Assert{
			{Field: "id", Kind: LITERAL},
			{Field: "name", Kind: STRING},
			{Field: "age", Kind: LITERAL},
			{Field: "category", Kind: STRING},
		})
	g.Expect(err).To(gomega.BeNil())

	p = Parser{}
	filter, err = p.Filter("name:elmer,age:20")
	g.Expect(err).To(gomega.BeNil())
	err = filter.Validate(
		[]Assert{
			{Field: "id", Kind: LITERAL},
			{Field: "name", Kind: STRING},
		})
	g.Expect(err).ToNot(gomega.BeNil())

	p = Parser{}
	filter, err = p.Filter("age~10")
	g.Expect(err).To(gomega.BeNil())
	err = filter.Validate(
		[]Assert{
			{Field: "age", Kind: LITERAL},
		})
	g.Expect(err).ToNot(gomega.BeNil())
}

func TestFieldSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	field := func(name string) (f *Field) {
		f = &Field{}
		f.Field.Value = name
		return
	}
	selector := FieldSelector{
		"zero",
		"-one",
		"-two",
	}
	g.Expect(selector.Match(field("zero"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("one"))).ToNot(gomega.BeTrue())
	g.Expect(selector.Match(field("two"))).ToNot(gomega.BeTrue())
	g.Expect(selector.Match(field("unknown"))).ToNot(gomega.BeTrue())

	selector = FieldSelector{
		"-one",
		"-two",
	}
	g.Expect(selector.Match(field("Zero"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("unknown"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("one"))).ToNot(gomega.BeTrue())
	g.Expect(selector.Match(field("two"))).ToNot(gomega.BeTrue())

	selector = FieldSelector{
		"one",
		"two",
	}
	g.Expect(selector.Match(field("one"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("two"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("unknown"))).ToNot(gomega.BeTrue())

	selector = FieldSelector{}
	g.Expect(selector.Match(field("one"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("two"))).To(gomega.BeTrue())

	selector = FieldSelector{
		"resource.one",
		"two",
	}
	g.Expect(selector.Match(field("two"))).To(gomega.BeTrue())
	g.Expect(selector.Match(field("resource.one"))).ToNot(gomega.BeTrue())
}

func TestFilterWith(t *testing.T) {
	var err error
	g := gomega.NewGomegaWithT(t)
	p := Parser{}

	filter, err := p.Filter("name:elmer,age:20,category=(a|b|c)")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(filter.predicates)).To(gomega.Equal(3))
	g.Expect(filter.Empty()).To(gomega.BeFalse())

	f := filter.With("-name")
	_, hasName := f.Field("name")
	g.Expect(len(f.predicates)).To(gomega.Equal(2))
	g.Expect(hasName).To(gomega.BeFalse())

	f = filter.With("+name", "+age")
	_, hasName = f.Field("name")
	_, hasAge := f.Field("age")
	_, hasCat := f.Field("category")
	g.Expect(len(f.predicates)).To(gomega.Equal(2))
	g.Expect(hasAge).To(gomega.BeTrue())
	g.Expect(hasCat).To(gomega.BeFalse())
}
