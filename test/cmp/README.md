# CMP - Deep Object Comparison Package

Package `cmp` provides deep comparison of Go objects with detailed difference reporting. It uses reflection to traverse and compare complex data structures including structs, slices, maps, pointers, and primitives.

## Features

- **Deep comparison** of any Go types
- **Detailed diff reports** showing exactly what differs
- **Path-based ignore** to skip specific fields
- **Support for all common types**: primitives, strings, slices, maps, structs, pointers, time.Time
- **Nested structure support** with dot notation paths
- **Anonymous field handling** in structs
- **Nil-safe** comparisons

## API

### Eq(expected, got any, ignoredPaths ...string) (eq bool, report string)

The primary comparison function that returns both a boolean result and a formatted report.

```go
eq, report := cmp.Eq(structA, structB)
if !eq {
    fmt.Println(report)
}
```

### Inspect(a, b any, ignoredPaths ...string) (eq bool, diff string)

Returns just the diff section without the full "Expected/Got" formatting.

```go
eq, diff := cmp.Inspect(objA, objB)
```

### New(ignoredPaths ...string) *EQ

Creates a reusable comparator with configured ignore paths.

```go
cmp := cmp.New("Created", "UpdatedAt")
eq, report := cmp.Is(structA, structB)
```

## Report Format

When objects differ, `Eq()` returns a detailed report with three sections:

### 1. Expected Section
Shows the formatted representation of the expected value (first argument).

### 2. Got Section
Shows the formatted representation of the actual value (second argument).

### 3. Diff Section
Lists all differences with the following symbols:

- `~` - **Modified**: field value differs between objects
- `+` - **Added**: field exists in Got but not in Expected
- `-` - **Removed**: field exists in Expected but not in Got

## Examples

### Basic Comparison

```go
eq, report := cmp.Eq(42, 100)
// eq = false
// report contains: ~  = 100 expected: 42
```

### String Comparison

```go
eq, report := cmp.Eq("hello", "world")
// Diff:
// ~  = "world" expected: "hello"
```

### Type Mismatch

```go
eq, report := cmp.Eq(42, "42")
// Diff:
// : (type) int != string
```

### Slice Comparison

```go
listA := []int{1, 2, 3}
listB := []int{1, 3, 2}

eq, report := cmp.Eq(listA, listB)
// Diff:
// ~ [1] = 3 expected: 2
// ~ [2] = 2 expected: 3
```

### Slice Length Differences

```go
listA := []int{1, 2, 3, 4}
listB := []int{1, 3, 2}

eq, report := cmp.Eq(listA, listB)
// Diff:
// ~ [1] = 3 expected: 2
// ~ [2] = 2 expected: 3
// - [3] = 4
```

### Struct Comparison

```go
type Person struct {
    Name    string
    Age     int
    Address Address
}

type Address struct {
    City  string
    State string
}

personA := Person{
    Name: "Alice",
    Age:  30,
    Address: Address{
        City:  "Boston",
        State: "MA",
    },
}

personB := Person{
    Name: "Bob",
    Age:  30,
    Address: Address{
        City:  "Seattle",
        State: "WA",
    },
}

eq, report := cmp.Eq(personA, personB)
// Diff:
// ~ Name = "Bob" expected: "Alice"
// ~ Address.City = "Seattle" expected: "Boston"
// ~ Address.State = "WA" expected: "MA"
```

### Map Comparison

```go
mapA := map[string]int{"a": 1, "b": 2}
mapB := map[string]int{"a": 1, "b": 3}

eq, report := cmp.Eq(mapA, mapB)
// Diff:
// ~ b: 3 expected: 2
```

### Map with Extra Keys

```go
mapA := map[string]int{"a": 1, "b": 2}
mapB := map[string]int{"a": 1}

eq, report := cmp.Eq(mapA, mapB)
// Diff:
// ~ : b<ptr> expected: <nil>
```

### Pointer Comparison

```go
type Data struct {
    Value string
}

a := &Data{Value: "John"}
b := &Data{Value: "Jane"}

eq, report := cmp.Eq(a, b)
// Diff:
// ~ Value = "Jane" expected: "John"
```

### Nil Pointer Comparison

```go
a := &Data{Value: "John"}
var b *Data = nil

eq, report := cmp.Eq(a, b)
// Diff:
// ~  = <ptr> expected: <nil>
```

## Ignoring Fields

Use the ignore paths feature to skip specific fields during comparison. Paths use dot notation to traverse nested structures.

### Ignore Single Field

```go
type Record struct {
    ID      int
    Created time.Time
    Name    string
}

a := Record{ID: 1, Created: time.Now(), Name: "Alice"}
b := Record{ID: 1, Created: time.Now().Add(time.Hour), Name: "Alice"}

eq, report := cmp.Eq(a, b, "Created")
// eq = true (Created field is ignored)
```

### Ignore Nested Fields

```go
type User struct {
    Name    string
    Profile *Profile
}

type Profile struct {
    Bio     string
    Created time.Time
}

eq, report := cmp.Eq(userA, userB, "Profile.Created")
// Ignores only the Created field within Profile
```

### Ignore Entire Nested Object

```go
eq, report := cmp.Eq(userA, userB, "Profile")
// Ignores the entire Profile object
```

### Multiple Ignore Paths

```go
eq, report := cmp.Eq(structA, structB, "Created", "UpdatedAt", "Profile.Timestamp")
```

## Supported Types

- **Primitives**: int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool
- **Strings**: full string comparison with length information
- **Slices**: element-by-element comparison with index tracking
- **Arrays**: similar to slices
- **Maps**: key-by-key comparison with sorted output
- **Structs**: field-by-field comparison (exported fields only)
- **Pointers**: dereferences and compares underlying values
- **time.Time**: uses `time.Equal()` for proper timezone handling
- **Interfaces**: compares underlying concrete types
- **Nil values**: properly handles nil pointers, slices, and maps

## Special Behaviors

### Nil vs Empty Slices/Maps

Nil slices and empty slices are considered equal:

```go
eq, _ := cmp.Eq([]int(nil), []int{})
// eq = true
```

### Unexported Fields

Unexported (private) struct fields are ignored during comparison:

```go
type Data struct {
    Public  string
    private string  // ignored
}
```

### Anonymous (Embedded) Fields

Anonymous struct fields are traversed without adding an extra path component:

```go
type Name struct {
    First string
    Last  string
}

type Person struct {
    Name  // embedded
    Age int
}

// Differences reported as "First", "Last", "Age"
// not "Name.First", "Name.Last"
```

### Time Comparison

`time.Time` values are compared using `time.Equal()`, which correctly handles different timezones representing the same instant.

## Testing

The package includes comprehensive tests covering:

- Primitive types (int, float, bool, string)
- Type mismatches
- Slices (equal, different order, different length)
- Nil comparisons
- Pointer comparisons
- Map comparisons (equal, different values, extra keys)
- Time comparisons
- Struct comparisons with nested objects
- Ignore path functionality
- Anonymous fields
- Unexported fields

Run tests:

```bash
go test ./test/cmp
```

Run tests with report output:

```bash
go test -v ./test/cmp
```

## Use Cases

- **Unit testing**: Compare expected vs actual objects with detailed failure reports
- **API testing**: Verify response objects match expectations
- **Data validation**: Compare database records before/after operations
- **Integration testing**: Validate complex object transformations
- **Debugging**: Understand exactly how two objects differ

## Implementation Notes

- Uses `reflect` package for deep traversal
- Uses `davecgh/go-spew` for formatted output
- Sorts map keys for deterministic output
- Maintains path context during traversal for precise error reporting
- Memory efficient for large structures (streams comparison results)
