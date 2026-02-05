// This basically has enough examples to cover most use-cases of the lib
package main

import (
	"fmt"
	"strictjson"
)

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZipCode string `json:"zipCode"`
	Country string `json:"country"`
}

type ContactInfo struct {
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Address Address `json:"address"`
}

type Department struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	IsActive bool   `json:"isActive"`
}

type Employee struct {
	ID          int              `json:"id"`
	FirstName   string           `json:"firstName"`
	LastName    string           `json:"lastName"`
	Contact     ContactInfo      `json:"contact"`
	Departments []Department     `json:"departments"`
	Metadata    map[string]Skill `json:"metadata"`
}

type Skill struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	IsCertified bool   `json:"isCertified"`
}

func main() {
	// Example 1: Valid JSON with correct case
	validJSON := []byte(`{
		"id": 1,
		"firstName": "John",
		"lastName": "Doe",
		"contact": {
			"email": "john.doe@example.com",
			"phone": "+1-555-0123",
			"address": {
				"street": "123 Main St",
				"city": "New York",
				"zipCode": "10001",
				"country": "USA"
			}
		},
		"departments": [
			{"name": "Engineering", "code": "ENG", "isActive": true},
			{"name": "Research", "code": "RND", "isActive": true}
		],
		"metadata": {
			"primary": {"name": "Go", "level": 5, "isCertified": true},
			"secondary": {"name": "Python", "level": 4, "isCertified": false}
		}
	}`)

	fmt.Println("Example 1: Valid JSON with correct case")
	fmt.Println("-------------------------------------------")
	var emp1 Employee
	if err := strictjson.Unmarshal(validJSON, &emp1); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✓ Success! Parsed employee: %s %s\n", emp1.FirstName, emp1.LastName)
		fmt.Printf("  Contact: %s, %s\n", emp1.Contact.Email, emp1.Contact.Address.City)
		fmt.Printf("  Departments: %d\n", len(emp1.Departments))
		fmt.Printf("  Skills: %d\n", len(emp1.Metadata))
	}
	fmt.Println()

	// Example 2: Invalid JSON - wrong case in nested struct (Address)
	invalidNestedJSON := []byte(`{
		"id": 2,
		"firstName": "Jane",
		"lastName": "Smith",
		"contact": {
			"email": "jane.smith@example.com",
			"phone": "+1-555-0456",
			"address": {
				"street": "456 Oak Ave",
				"CITY": "Boston",
				"zipCode": "02101",
				"country": "USA"
			}
		},
		"departments": [],
		"metadata": {}
	}`)

	fmt.Println("Example 2: Invalid JSON - wrong case in nested struct ('CITY' vs 'city')")
	fmt.Println("-------------------------------------------------------------------------")
	var emp2 Employee
	if err := strictjson.Unmarshal(invalidNestedJSON, &emp2); err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	} else {
		fmt.Println("Unexpectedly succeeded!")
	}
	fmt.Println()

	// Example 3: Invalid JSON - wrong case in slice element
	invalidSliceJSON := []byte(`{
		"id": 3,
		"firstName": "Bob",
		"lastName": "Johnson",
		"contact": {
			"email": "bob@example.com",
			"phone": "+1-555-0789",
			"address": {
				"street": "789 Pine Rd",
				"city": "Chicago",
				"zipCode": "60601",
				"country": "USA"
			}
		},
		"departments": [
			{"name": "Sales", "code": "SAL", "isActive": true},
			{"name": "Marketing", "Code": "MKT", "isActive": false}
		],
		"metadata": {}
	}`)

	fmt.Println("Example 3: Invalid JSON - wrong case in slice element ('Code' vs 'code')")
	fmt.Println("-------------------------------------------------------------------------")
	var emp3 Employee
	if err := strictjson.Unmarshal(invalidSliceJSON, &emp3); err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	} else {
		fmt.Println("Unexpectedly succeeded!")
	}
	fmt.Println()

	// Example 4: Invalid JSON - wrong case in map value struct
	invalidMapJSON := []byte(`{
		"id": 4,
		"firstName": "Alice",
		"lastName": "Williams",
		"contact": {
			"email": "alice@example.com",
			"phone": "+1-555-1234",
			"address": {
				"street": "321 Elm St",
				"city": "Seattle",
				"zipCode": "98101",
				"country": "USA"
			}
		},
		"departments": [],
		"metadata": {
			"main": {"name": "JavaScript", "level": 4, "IsCertified": true}
		}
	}`)

	fmt.Println("Example 4: Invalid JSON - wrong case in map value ('IsCertified' vs 'isCertified')")
	fmt.Println("-----------------------------------------------------------------------------------")
	var emp4 Employee
	if err := strictjson.Unmarshal(invalidMapJSON, &emp4); err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	} else {
		fmt.Println("Unexpectedly succeeded!")
	}
	fmt.Println()

	// Example 5: Using decoder with suggestions enabled
	fmt.Println("Example 5: Using decoder with 'did you mean?' suggestions")
	fmt.Println("-----------------------------------------------------------")
	wrongCaseJSON := []byte(`{
		"id": 5,
		"FirstName": "Charlie",
		"lastName": "Brown",
		"contact": {
			"email": "charlie@example.com",
			"phone": "+1-555-5678",
			"address": {
				"street": "555 Maple Dr",
				"city": "Denver",
				"zipCode": "80201",
				"country": "USA"
			}
		},
		"departments": [],
		"metadata": {}
	}`)

	decoder := strictjson.NewDecoder(strictjson.WithSuggestClosest(true))
	var emp5 Employee
	if err := decoder.Unmarshal(wrongCaseJSON, &emp5); err != nil {
		fmt.Printf("Error with suggestion: %v\n", err)
	} else {
		fmt.Println("Unexpectedly succeeded!")
	}
	fmt.Println()

	// Example 6: Deeply nested structure validation
	type Inner struct {
		Value string `json:"value"`
	}
	type Middle struct {
		Inner Inner `json:"inner"`
	}
	type Outer struct {
		Middle Middle `json:"middle"`
	}

	fmt.Println("Example 6: Deeply nested struct validation")
	fmt.Println("-------------------------------------------")
	deeplyNestedInvalidJSON := []byte(`{
		"middle": {
			"inner": {
				"VALUE": "test"
			}
		}
	}`)

	var outer Outer
	if err := strictjson.Unmarshal(deeplyNestedInvalidJSON, &outer); err != nil {
		fmt.Printf("Error at deepest level ('VALUE' vs 'value'): %v\n", err)
	} else {
		fmt.Println("Unexpectedly succeeded!")
	}
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}
