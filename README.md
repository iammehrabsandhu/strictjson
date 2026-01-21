# strictjson

`strictjson` is a Go library for **case-sensitive JSON unmarshalling**. It validates that JSON keys exactly match struct field names (or their `json` tags) before unmarshalling.

Unlike the standard `encoding/json` package which uses case-insensitive matching, `strictjson` enforces strict casing rules, making it ideal for APIs with strict contracts or when precise field validation is required.

## Features

- **Case-Sensitive Validation**: Enforces exact case matching for struct fields.
- **Recursive Validation**: Validates nested structs, slices of structs, and maps with struct values.
- **Helpful Errors**: Reports specific unknown fields and offers "did you mean?" suggestions.
- **Configurable**: Options to allow/disallow unknown fields and enable/disable suggestions.
- **Standard Compatible**: APIs mirror `encoding/json` for easy drop-in replacement.
- **Custom Unmarshaler Support**: Respects types that implement `json.Unmarshaler`.

## Installation

```bash
go get github.com/yourusername/strictjson
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"strictjson"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// Valid JSON
	json1 := []byte(`{"name": "John", "age": 30}`)
	var p1 Person
	if err := strictjson.Unmarshal(json1, &p1); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Success: %+v\n", p1)

	// Invalid JSON (wrong case)
	json2 := []byte(`{"Name": "John", "age": 30}`)
	var p2 Person
	if err := strictjson.Unmarshal(json2, &p2); err != nil {
		fmt.Println("Error:", err) 
		// Output: Error: strictjson: unknown field "Name"
	}
}
```

### Validator Options

You can configure the decoder behavior:

```go
// Allow unknown fields (only validate matching fields are correct case)
d := strictjson.NewDecoder(strictjson.WithDisallowUnknownFields(false))

// Enable suggestions for unknown fields
d := strictjson.NewDecoder(strictjson.WithSuggestClosest(true))
// Error: strictjson: unknown field "Name" (did you mean "name"?)
```

### Recursive Validation

`strictjson` automatically validates nested structures:

```go
type Config struct {
    Database struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"database"`
}

// This will fail because "HOST" does not match "host"
data := []byte(`{
    "database": {
        "HOST": "localhost", 
        "port": 5432
    }
}`)
```

## Performance

`strictjson` uses reflection to traverse and validate the structure before delegating to `encoding/json` for the actual parsing. This adds some overhead (approx 2.5x slower than standard library), but provides strict validation guarantees essential for robust API integrations.
