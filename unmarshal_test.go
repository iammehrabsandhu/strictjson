package strictjson

import (
	"encoding/json"
	"testing"
	"time"
)

// =============================================================================
// Basic Case-Sensitive Matching Tests
// =============================================================================

func TestBasicCaseSensitive(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "exact case match",
			json:    `{"name": "John", "age": 30}`,
			wantErr: false,
		},
		{
			name:    "wrong case Name",
			json:    `{"Name": "John", "age": 30}`,
			wantErr: true,
		},
		{
			name:    "wrong case AGE",
			json:    `{"name": "John", "AGE": 30}`,
			wantErr: true,
		},
		{
			name:    "all wrong case",
			json:    `{"NAME": "John", "AGE": 30}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Person
			err := Unmarshal([]byte(tt.json), &p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && p.Name != "John" {
				t.Errorf("Expected Name='John', got '%s'", p.Name)
			}
		})
	}
}

func TestFieldNameVsTag(t *testing.T) {
	type Config struct {
		ServerURL string `json:"server_url"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "matches tag",
			json:    `{"server_url": "http://example.com"}`,
			wantErr: false,
		},
		{
			name:    "matches Go field name - should fail",
			json:    `{"ServerURL": "http://example.com"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Config
			err := Unmarshal([]byte(tt.json), &c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Nested Struct Validation Tests
// =============================================================================

func TestNestedStructCaseSensitive(t *testing.T) {
	type Address struct {
		City    string `json:"city"`
		ZipCode string `json:"zipCode"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "all correct case",
			json:    `{"name": "John", "address": {"city": "NYC", "zipCode": "10001"}}`,
			wantErr: false,
		},
		{
			name:    "nested field wrong case",
			json:    `{"name": "John", "address": {"CITY": "NYC", "zipCode": "10001"}}`,
			wantErr: true,
		},
		{
			name:    "nested field wrong case - zipcode",
			json:    `{"name": "John", "address": {"city": "NYC", "zipcode": "10001"}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Person
			err := Unmarshal([]byte(tt.json), &p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeeplyNestedStruct(t *testing.T) {
	type Inner struct {
		Value string `json:"value"`
	}
	type Middle struct {
		Inner Inner `json:"inner"`
	}
	type Outer struct {
		Middle Middle `json:"middle"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "all correct",
			json:    `{"middle": {"inner": {"value": "test"}}}`,
			wantErr: false,
		},
		{
			name:    "deepest level wrong case",
			json:    `{"middle": {"inner": {"VALUE": "test"}}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var o Outer
			err := Unmarshal([]byte(tt.json), &o)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Slice/Array Tests
// =============================================================================

func TestSliceOfStructs(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
		wantLen int
	}{
		{
			name:    "all correct case",
			json:    `[{"name": "Apple", "price": 100}, {"name": "Banana", "price": 50}]`,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "one item wrong case",
			json:    `[{"name": "Apple", "price": 100}, {"NAME": "Banana", "price": 50}]`,
			wantErr: true,
		},
		{
			name:    "empty array",
			json:    `[]`,
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var items []Item
			err := Unmarshal([]byte(tt.json), &items)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(items) != tt.wantLen {
				t.Errorf("Expected %d items, got %d", tt.wantLen, len(items))
			}
		})
	}
}

func TestSliceOfPrimitives(t *testing.T) {
	// For primitives, no case validation needed
	var nums []int
	err := Unmarshal([]byte(`[1, 2, 3]`), &nums)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error = %v", err)
	}
	if len(nums) != 3 {
		t.Errorf("Expected 3 items, got %d", len(nums))
	}
}

// =============================================================================
// Map Tests
// =============================================================================

func TestMapWithStructValues(t *testing.T) {
	type Config struct {
		Value   string `json:"value"`
		Enabled bool   `json:"enabled"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "all correct case",
			json:    `{"key1": {"value": "a", "enabled": true}, "key2": {"value": "b", "enabled": false}}`,
			wantErr: false,
		},
		{
			name:    "one value wrong case",
			json:    `{"key1": {"value": "a", "enabled": true}, "key2": {"VALUE": "b", "enabled": false}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m map[string]Config
			err := Unmarshal([]byte(tt.json), &m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMapWithPrimitiveValues(t *testing.T) {
	// For primitive values, no struct validation
	var m map[string]int
	err := Unmarshal([]byte(`{"a": 1, "b": 2}`), &m)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error = %v", err)
	}
	if m["a"] != 1 || m["b"] != 2 {
		t.Errorf("Unexpected map values: %v", m)
	}
}

// =============================================================================
// Embedded Struct Tests
// =============================================================================

func TestEmbeddedStruct(t *testing.T) {
	type Base struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type Extended struct {
		Base
		Extra string `json:"extra"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "all correct case",
			json:    `{"id": 1, "name": "test", "extra": "value"}`,
			wantErr: false,
		},
		{
			name:    "embedded field wrong case",
			json:    `{"ID": 1, "name": "test", "extra": "value"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Extended
			err := Unmarshal([]byte(tt.json), &e)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmbeddedStructConflict(t *testing.T) {
	type A struct {
		Name string `json:"name"`
	}
	type B struct {
		Name string `json:"name"`
	}
	type Conflict struct {
		A
		B
	}

	var c Conflict
	err := Unmarshal([]byte(`{"name": "test"}`), &c)
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}
}

// =============================================================================
// Pointer Field Tests
// =============================================================================

func TestPointerFields(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}
	type Person struct {
		Name    string   `json:"name"`
		Address *Address `json:"address"`
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "pointer field with correct case",
			json:    `{"name": "John", "address": {"city": "NYC"}}`,
			wantErr: false,
		},
		{
			name:    "pointer field with wrong case",
			json:    `{"name": "John", "address": {"CITY": "NYC"}}`,
			wantErr: true,
		},
		{
			name:    "null pointer field",
			json:    `{"name": "John", "address": null}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Person
			err := Unmarshal([]byte(tt.json), &p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Custom Unmarshaler Tests
// =============================================================================

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// Custom unmarshaling logic - accepts any format
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

func TestCustomUnmarshaler(t *testing.T) {
	type Event struct {
		Name string     `json:"name"`
		Date CustomTime `json:"date"`
	}

	// Custom unmarshaler should be respected
	var e Event
	err := Unmarshal([]byte(`{"name": "Birthday", "date": "2024-01-15"}`), &e)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error = %v", err)
	}
	if e.Name != "Birthday" {
		t.Errorf("Expected Name='Birthday', got '%s'", e.Name)
	}
}

func TestStructImplementsUnmarshaler(t *testing.T) {
	// If the entire struct implements Unmarshaler, we delegate completely
	var ct CustomTime
	err := Unmarshal([]byte(`"2024-01-15"`), &ct)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error = %v", err)
	}
}

// =============================================================================
// Decoder Options Tests
// =============================================================================

func TestDisallowUnknownFieldsOption(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	// Default: DisallowUnknownFields = true
	var p1 Person
	err := Unmarshal([]byte(`{"name": "John", "extra": "field"}`), &p1)
	if err == nil {
		t.Error("Expected error for unknown field with default settings")
	}

	// With DisallowUnknownFields = false
	d := NewDecoder(WithDisallowUnknownFields(false))
	var p2 Person
	err = d.Unmarshal([]byte(`{"name": "John", "extra": "field"}`), &p2)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error with DisallowUnknownFields=false: %v", err)
	}
	if p2.Name != "John" {
		t.Errorf("Expected Name='John', got '%s'", p2.Name)
	}
}

func TestSuggestClosestOption(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	// With SuggestClosest = true
	d := NewDecoder(WithSuggestClosest(true))
	var p Person
	err := d.Unmarshal([]byte(`{"Name": "John"}`), &p)
	if err == nil {
		t.Error("Expected error for wrong case field")
	}
	if err != nil && err.Error() != `strictjson: unknown field "Name" (did you mean "name"?)` {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEmptyJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	var p Person
	err := Unmarshal([]byte(`{}`), &p)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error: %v", err)
	}
	if p.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", p.Name)
	}
}

func TestNullJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	var p Person
	err := Unmarshal([]byte(`null`), &p)
	if err != nil {
		t.Errorf("Unmarshal() unexpected error: %v", err)
	}
}

func TestMalformedJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	var p Person
	err := Unmarshal([]byte(`{invalid}`), &p)
	if err == nil {
		t.Error("Expected error for malformed JSON")
	}
}

func TestNonPointerTarget(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	var p Person
	err := Unmarshal([]byte(`{"name": "John"}`), p) // Note: not a pointer
	if err == nil {
		t.Error("Expected error for non-pointer target")
	}
}

func TestNilPointerTarget(t *testing.T) {
	var p *struct{ Name string }
	err := Unmarshal([]byte(`{"name": "John"}`), p)
	if err == nil {
		t.Error("Expected error for nil pointer target")
	}
}

func TestJsonTagDash(t *testing.T) {
	type Person struct {
		Name   string `json:"name"`
		Secret string `json:"-"`
	}

	var p Person
	err := Unmarshal([]byte(`{"name": "John", "Secret": "hidden"}`), &p)
	// "Secret" with json:"-" should be treated as unknown
	if err == nil {
		t.Error("Expected error for field with json:\"-\" tag")
	}
}

// =============================================================================
// Real World Scenario Tests
// =============================================================================

func TestRealWorldScenario(t *testing.T) {
	// A complex structure mimicking a cloud provider API response
	type ResourceTags struct {
		Environment string `json:"Environment"`
		Owner       string `json:"Owner"`
	}

	type InstanceMetrics struct {
		CPUUsage    float64 `json:"CpuUsage"`
		MemoryUsage float64 `json:"MemoryUsage"`
	}

	type VMInstance struct {
		ID        string            `json:"InstanceId"`
		Type      string            `json:"InstanceType"`
		Launched  time.Time         `json:"LaunchTime"`
		Tags      ResourceTags      `json:"Tags"`
		Metrics   []InstanceMetrics `json:"Metrics"`
		IsRunning bool              `json:"IsRunning"`
	}

	type RegionConfig struct {
		RegionName string                `json:"Region"`
		Instances  map[string]VMInstance `json:"Instances"`
	}

	type CloudResponse struct {
		RequestID string         `json:"RequestId"`
		Status    int            `json:"Status"`
		Data      []RegionConfig `json:"Data"`
	}

	// Valid JSON matches exactly
	validJSON := `{
		"RequestId": "req-123",
		"Status": 200,
		"Data": [
			{
				"Region": "us-east-1",
				"Instances": {
					"vm-1": {
						"InstanceId": "i-1234567890abcdef0",
						"InstanceType": "t3.micro",
						"LaunchTime": "2024-01-20T10:00:00Z",
						"Tags": {
							"Environment": "Production",
							"Owner": "DevOps"
						},
						"Metrics": [
							{"CpuUsage": 15.5, "MemoryUsage": 42.0}
						],
						"IsRunning": true
					}
				}
			}
		]
	}`

	var resp CloudResponse
	if err := Unmarshal([]byte(validJSON), &resp); err != nil {
		t.Errorf("Unmarshal() valid JSON failed: %v", err)
	}

	// Validate deep structure
	if len(resp.Data) != 1 || resp.Data[0].RegionName != "us-east-1" {
		t.Error("Failed to unmarshal deep structure correctly")
	}
	vm := resp.Data[0].Instances["vm-1"]
	if vm.Tags.Environment != "Production" {
		t.Error("Failed to unmarshal nested struct fields (Tags)")
	}

	// Invalid JSON: Case mismatch deep in the structure
	// "cpuUsage" instead of "CpuUsage"
	invalidJSON := `{
		"RequestId": "req-123",
		"Status": 200,
		"Data": [
			{
				"Region": "us-east-1",
				"Instances": {
					"vm-1": {
						"InstanceId": "i-xxx",
						"InstanceType": "t3.micro",
						"LaunchTime": "2024-01-20T10:00:00Z",
						"Tags": {"Environment": "Dev", "Owner": "Me"},
						"Metrics": [
							{"cpuUsage": 15.5, "MemoryUsage": 42.0}
						],
						"IsRunning": true
					}
				}
			}
		]
	}`

	var invalidResp CloudResponse
	err := Unmarshal([]byte(invalidJSON), &invalidResp)
	if err == nil {
		t.Error("Expected error for case mismatch in deep nested slice struct (cpuUsage vs CpuUsage)")
	}
}

func TestComplexNightmareScenario(t *testing.T) {
	// A recursive, deeply nested structure with mixed types
	type MetaData struct {
		CreatedAt time.Time         `json:"createdAt"`
		Tags      map[string]string `json:"tags"`
		Priority  *int              `json:"priority"`
	}

	type ConfigNode struct {
		Key       string    `json:"key"`
		Value     any       `json:"val"` // Interface shouldn't be recursed by strictjson, but let's test surrounding bits
		IsEnabled bool      `json:"isEnabled"`
		Meta      *MetaData `json:"meta"`
	}

	type deeplyRecursiveNode struct {
		ID       string                         `json:"id"`
		Children []deeplyRecursiveNode          `json:"children"` // Recursive slice
		Config   map[string]*ConfigNode         `json:"config"`   // Map of pointers
		Siblings map[string]deeplyRecursiveNode `json:"siblings"` // Recursive map
		Next     *deeplyRecursiveNode           `json:"next"`     // Recursive pointer
	}

	type RootPayload struct {
		Version string              `json:"version"`
		Graph   deeplyRecursiveNode `json:"graph"`
	}

	// 1. Valid deep nesting
	validJSON := `{
		"version": "1.0",
		"graph": {
			"id": "root",
			"children": [
				{
					"id": "child_1",
					"children": [],
					"config": {
						"main": {
							"key": "timeout",
							"val": 100,
							"isEnabled": true,
							"meta": {
								"createdAt": "2024-01-01T00:00:00Z",
								"tags": {"env": "prod"},
								"priority": 1
							}
						}
					}
				}
			],
			"next": {
				"id": "root_next",
				"siblings": {
					"alien": {
						"id": "alien_1",
						"next": {
							"id": "alien_next",
							"config": {
								"sub": {
									"key": "sub_key",
									"isEnabled": false,
									"meta": {
										"createdAt": "2024-01-01T00:00:00Z",
										"tags": {},
										"priority": 2
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	var root RootPayload
	if err := Unmarshal([]byte(validJSON), &root); err != nil {
		t.Fatalf("Unmarshal() failed on valid deep structure: %v", err)
	}

	// Verify deep data extraction
	if root.Graph.Next.Siblings["alien"].Next.Config["sub"].Meta.Priority == nil {
		t.Fatal("Failed to unmarshal deeply nested pointer value")
	}
	if *root.Graph.Next.Siblings["alien"].Next.Config["sub"].Meta.Priority != 2 {
		t.Errorf("Expected deeply nested priority 2, got %d", *root.Graph.Next.Siblings["alien"].Next.Config["sub"].Meta.Priority)
	}

	// 2. Invalid: Deep nested casing error (8 levels deep)
	// Changing "priority" to "Priority" inside the deepest node
	invalidJSON := `{
		"version": "1.0",
		"graph": {
			"id": "root",
			"next": {
				"id": "root_next",
				"siblings": {
					"alien": {
						"id": "alien_1",
						"next": {
							"id": "alien_next",
							"config": {
								"sub": {
									"key": "sub_key",
									"isEnabled": false,
									"meta": {
										"createdAt": "2024-01-01T00:00:00Z",
										"tags": {},
										"Priority": 2
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	var rootInvalid RootPayload
	err := Unmarshal([]byte(invalidJSON), &rootInvalid)
	if err == nil {
		t.Error("Did not detect case mismatch 'Priority' at 8th level of nesting!")
	} else {
		expectedSubstr := `strictjson: unknown or mis-cased field "Priority"`
		if !contains(err.Error(), expectedSubstr) {
			t.Errorf("Expected error containing %q, got %q", expectedSubstr, err.Error())
		}
	}

	// 3. Invalid: Error inside map key struct (3 levels deep)
	// "isEnabled" -> "IsEnabled"
	invalidJSON2 := `{
		"version": "1.0",
		"graph": {
			"id": "root",
			"children": [
				{
					"id": "child_1",
					"config": {
						"main": {
							"key": "timeout",
							"val": 100,
							"IsEnabled": true, 
							"meta": null
						}
					}
				}
			]
		}
	}`

	err = Unmarshal([]byte(invalidJSON2), &rootInvalid)
	if err == nil {
		t.Error("Did not detect case mismatch 'IsEnabled' inside map of pointers!")
	} else {
		expectedSubstr := `strictjson: unknown or mis-cased field "IsEnabled"`
		if !contains(err.Error(), expectedSubstr) {
			t.Errorf("Expected error containing %q, got %q", expectedSubstr, err.Error())
		}
	}
}

// Helper for error string checking
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr ||
		(len(s) > len(substr) && contains(s[1:], substr))
}

// Using strings.Contains would be better but trying to avoid adding imports if possible.
// Actually testing package implies we can import strings if not present.
// Let's rely on standard strings.Contains if we add the import, or just a simple check.
// The file imports "strings" already? Let me check imports.
// Checking file imports... imports "encoding/json", "testing", "time".
// I will just use a simple helper or add "strings" to imports.
// To be safe and clean, I'll update imports in a separate step if needed, or just write a simple loop.
// Actually, `strings` is likely needed for robust testing. Let me check imports first.

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkUnmarshalSimple(b *testing.B) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := []byte(`{"name": "John", "age": 30}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		_ = Unmarshal(data, &p)
	}
}

func BenchmarkUnmarshalNested(b *testing.B) {
	type Address struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	data := []byte(`{"name": "John", "address": {"city": "NYC", "country": "USA"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		_ = Unmarshal(data, &p)
	}
}

func BenchmarkUnmarshalSlice(b *testing.B) {
	type Item struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}

	data := []byte(`[{"name": "A", "price": 1}, {"name": "B", "price": 2}, {"name": "C", "price": 3}]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var items []Item
		_ = Unmarshal(data, &items)
	}
}

func BenchmarkStdlibUnmarshalSimple(b *testing.B) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := []byte(`{"name": "John", "age": 30}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		_ = json.Unmarshal(data, &p)
	}
}
