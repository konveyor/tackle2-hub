package filter

/*
?filter=predicates.  Each URL may contain multiple filter parameters.

Predicates:

filter (predicate (AND|OR)*)
predicate: field operator value
field: LITERAL
value: (LITERAL|STRING|list)
STRING: ('|")(.)*('|")
list: `(` (LITERAL|STRING) OR* `)` `
operator:
- ,  COMMA, AND
- | OR
- (:|,)= equal
- != not equal
- \> greater then
- \< less than
- \>= greater then or equal
- \<= less than or equal
- \\ escape quote
- ~ like
- () one of

\* is a wildcard for string matching.

Notes:
- The OR `|` operator may only be used inside a list `()`.


Examples:
?filter=name:elmer
?filter=name:"elmer"
?filter=name:'elmer'
?filter=name='elmer'
?filter=name!='elmer'
?filter=name~elmer*                   // name LIKE elmer%
?filter=category:(one|two)            // category IN (one, two)
?filter=category=(one|two)            // category IN (one, two)
?filter=category!=(one|two)           // category NOT IN (one, two)
?filter=category:mandatory,effort:20  // category=mandatory AND effort=20
?filter=category:mandatory|effort:10  // category=mandatory OR effort=10
?filter=tag.id:(1,2)                  // tag.id 1 AND 2.
*/
