## Overview

The `filter` package provides a query filtering mechanism that allows users to filter data based on specific conditions. 
The filter is applied using a query parameter `filter` in the URL.

## Syntax

The `filter` query parameter accepts a predicate expression that consists of one or more conditions combined using logical operators.

### Predicates

A predicate is a condition that is applied to a field. The general syntax of a predicate is:
```markdown
field operator value
```
* `field`: a literal string representing the field to be filtered.
* `operator`: one of the supported operators (see below).
* `value`: the value to be compared with the field. Can be a literal, string, or list.

### Operators

The following operators are supported:

| Operator | Description |
| --- | --- |
| `=`, `:`, `,` | Equal |
| `!=` | Not equal |
| `>` | Greater than |
| `<` | Less than |
| `>=` | Greater than or equal |
| `<=` | Less than or equal |
| `~` | Like (supports wildcard `*`) |
| `()` | One of (list) |

### List

A list is a collection of values enclosed in parentheses `()`. Values in the list are separated by the `OR` operator `|`.

### Logical Operators

* `AND` ( implicit or using comma `,` ): combines two or more conditions.
* `OR` ( `|` ): operator may only be used inside a list `()`.

## EBNF Grammar

```ebnf
Filter ::= Predicate ( ( "," | "|" ) Predicate )*
Predicate ::= Field Operator Value
Field ::= LITERAL
Operator ::= "=" | ":" | "!=" | ">" | "<" | ">=" | "<=" | "~"
Value ::= LITERAL | STRING | List
List ::= "(" ( LITERAL | STRING ) ( "|" ( LITERAL | STRING ) )* ")"
STRING ::= ( "'" | '"' ) ( . )* ( "'" | '"' )
LITERAL ::= [a-zA-Z0-9_]+
```

## Examples

#### Filter by name equal to "elmer".
?filter=name:elmer

#### Filter by name equal to "elmer" using a quoted string.
?filter=name:"elmer"

#### Filter by name not equal to "elmer".
?filter=name!='elmer'

#### Filter by name like "elmer%" using a wildcard.
?filter=name~elmer*

#### Filter by category in ("one", "two") using a list.
?filter=category:(one|two)

#### Filter by category in ("one", "two") using an equals operator.
?filter=category=(one|two)

#### Filter by category not in ("one", "two").
?filter=category!=(one|two)

#### Filter by category equal to "mandatory" and effort equal to 20.
?filter=category:mandatory,effort:20

#### Filter by effort greater than 0.
?filter=effort>0

## Notes

* The `OR` operator `|` can only be used inside a list `()`.
* The `*` wildcard is supported for string matching.
