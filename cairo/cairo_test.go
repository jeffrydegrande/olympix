package cairo_test

import (
	"testing"

	"github.com/jeffrydegrande/solidair/cairo"
)

func TestLanguage(t *testing.T) {
	lang := cairo.Language()
	if lang == nil {
		t.Errorf("Language() returned nil")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		source  []byte
		wantErr bool
	}{
		{
			name:    "Empty source",
			source:  []byte(""),
			wantErr: false,
		},
		{
			name: "Simple function",
			source: []byte(`
				func hello() {
					return "world";
				}
			`),
			wantErr: false,
		},
		{
			name: "Valid Cairo code",
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
				}
			`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := cairo.Parse(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tree == nil {
				t.Errorf("Parse() returned nil tree")
				return
			}
			
			// Verify tree has a root node
			root := tree.RootNode()
			if root == nil {
				t.Errorf("Parse() returned tree with nil root node")
				return
			}
		})
	}
}