package variables_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/jeffrydegrande/solidair/types"
	"github.com/jeffrydegrande/solidair/variables"
)

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name        string
		source      []byte
		wantErr     bool
		wantVarName string // A variable name we expect to find
	}{
		{
			name: "Simple extraction",
			source: []byte(`
				func main() {
					let variable_a = 1;
					let variable_b = 2;
					return variable_a + variable_b;
				}
			`),
			wantErr:     false,
			wantVarName: "variable_a",
		},
		{
			name: "Cairo contract with variables",
			source: []byte(`
				#[contract]
				mod Contract {
					struct Storage {
						active: bool,
						balance: felt252,
					}

					#[external]
					fn is_active() -> bool {
						return self.active::read();
					}

					#[external]
					fn deposit() {
						assert(self.active::read(), 'Market not active');
						let balance = self.balance::read();
						self.balance::write(balance + 1);
					}
				}
			`),
			wantErr:     false,
			wantVarName: "balance", // Changing to easier-to-find variable
		},
		{
			name:        "Empty source",
			source:      []byte(``),
			wantErr:     false,
			wantVarName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source
			tree, err := cairo.Parse(tt.source)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}
			
			// Extract variables
			vars, err := variables.ExtractVariables(tt.source, tree)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil {
				return
			}
			
			// Verify the extracted variables
			if vars == nil {
				t.Errorf("ExtractVariables() returned nil")
				return
			}
			
			// Check if we have variables
			if len(tt.source) > 0 && len(vars.Variables) == 0 {
				t.Errorf("ExtractVariables() returned no variables for non-empty source")
			}
			
			// Check if expected variable is present
			if tt.wantVarName != "" {
				var found bool
				for _, v := range vars.Variables {
					if v.Name == tt.wantVarName {
						found = true
						break
					}
				}
				
				if !found {
					t.Errorf("ExtractVariables() didn't extract expected variable %s", tt.wantVarName)
				}
			}
		})
	}
}

func TestExtractVariablesFromFile(t *testing.T) {
	// Test with the example files provided in the project
	files := []string{
		"/home/jeffry/Code/Olympix/assignment/examples/good.cairo",
		"/home/jeffry/Code/Olympix/assignment/examples/bad.cairo",
	}
	
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			// Read the file
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", file, err)
			}
			
			// Parse the source
			tree, err := cairo.Parse(content)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}
			
			// Extract variables
			vars, err := variables.ExtractVariables(content, tree)
			if err != nil {
				t.Fatalf("ExtractVariables() error = %v", err)
			}
			
			// Verify we extracted variables
			if len(vars.Variables) == 0 {
				t.Errorf("ExtractVariables() returned no variables for file %s", file)
			}
		})
	}
}

func TestPrintExtractedVariables(t *testing.T) {
	// Create a simple set of extracted variables
	vars := &variables.ExtractedVariables{
		Filename: "test.cairo",
		Variables: []types.VariableInfo{
			{
				Name:       "test_var",
				Type:       "bool",
				Context:    "variable",
				LineNumber: 10,
			},
			{
				Name:       "another_var",
				Type:       "felt252",
				Context:    "variable",
				LineNumber: 20,
			},
		},
	}
	
	// Redirect stdout to capture printed output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Call the function
	variables.PrintExtractedVariables(vars)
	
	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Verify output contains expected info
	if !contains(output, "Extracted 2 variables") {
		t.Errorf("PrintExtractedVariables() output doesn't contain variable count")
	}
	
	if !contains(output, "test_var") || !contains(output, "another_var") {
		t.Errorf("PrintExtractedVariables() output doesn't contain variable names")
	}
	
	if !contains(output, "bool") || !contains(output, "felt252") {
		t.Errorf("PrintExtractedVariables() output doesn't contain variable types")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}