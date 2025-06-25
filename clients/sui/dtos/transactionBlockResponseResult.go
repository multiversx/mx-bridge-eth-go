package dtos

// InspectResult represents the result of inspecting a transaction block
type InspectResult struct {
	ReturnValues [][]interface{} `json:"returnValues"`
}

// ReturnValue represents a single return value with its type
type ReturnValue struct {
	Value interface{}
	Type  string
}
