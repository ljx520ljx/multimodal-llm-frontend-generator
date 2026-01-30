package prompt

import (
	"testing"
)

func TestGetAllInteractionTypes(t *testing.T) {
	types := GetAllInteractionTypes()

	if len(types) != 6 {
		t.Errorf("Expected 6 interaction types, got %d", len(types))
	}

	expectedTypes := []InteractionType{
		InteractionToggle,
		InteractionExpand,
		InteractionTabSwitch,
		InteractionFormState,
		InteractionHover,
		InteractionNavigation,
	}

	for i, expected := range expectedTypes {
		if types[i] != expected {
			t.Errorf("Expected types[%d] = %q, got %q", i, expected, types[i])
		}
	}
}

func TestGetInteractionPattern(t *testing.T) {
	tests := []struct {
		interactionType InteractionType
		wantFound       bool
		wantStateType   string
		wantHandler     string
	}{
		{InteractionToggle, true, "boolean", "onClick"},
		{InteractionExpand, true, "boolean", "onClick"},
		{InteractionTabSwitch, true, "string", "onClick"},
		{InteractionFormState, true, "object", "onChange"},
		{InteractionHover, true, "boolean", "onMouseEnter"},
		{InteractionNavigation, true, "string", "onClick"},
		{InteractionType("invalid"), false, "", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.interactionType), func(t *testing.T) {
			pattern, found := GetInteractionPattern(tt.interactionType)

			if found != tt.wantFound {
				t.Errorf("GetInteractionPattern(%q) found = %v, want %v",
					tt.interactionType, found, tt.wantFound)
			}

			if found {
				if pattern.StateType != tt.wantStateType {
					t.Errorf("GetInteractionPattern(%q) StateType = %q, want %q",
						tt.interactionType, pattern.StateType, tt.wantStateType)
				}
				if pattern.EventHandler != tt.wantHandler {
					t.Errorf("GetInteractionPattern(%q) EventHandler = %q, want %q",
						tt.interactionType, pattern.EventHandler, tt.wantHandler)
				}
			}
		})
	}
}

func TestInteractionPatternsCompleteness(t *testing.T) {
	// Ensure every type in GetAllInteractionTypes has a pattern
	allTypes := GetAllInteractionTypes()

	for _, iType := range allTypes {
		pattern, found := GetInteractionPattern(iType)
		if !found {
			t.Errorf("Interaction type %q has no pattern defined", iType)
			continue
		}

		// Check pattern has required fields
		if pattern.Type != iType {
			t.Errorf("Pattern for %q has wrong Type: %q", iType, pattern.Type)
		}
		if pattern.Description == "" {
			t.Errorf("Pattern for %q has empty Description", iType)
		}
		if pattern.StateType == "" {
			t.Errorf("Pattern for %q has empty StateType", iType)
		}
		if pattern.EventHandler == "" {
			t.Errorf("Pattern for %q has empty EventHandler", iType)
		}
		if pattern.CodeExample == "" {
			t.Errorf("Pattern for %q has empty CodeExample", iType)
		}
	}
}

func TestInteractionTypeGuide(t *testing.T) {
	guide := InteractionTypeGuide()

	// Should be a markdown table
	if guide[0] != '|' {
		t.Error("InteractionTypeGuide should start with '|' (markdown table)")
	}

	// Should contain all type names
	expectedTypes := []string{"Toggle", "Expand", "TabSwitch", "FormState", "Hover", "Navigation"}
	for _, typeName := range expectedTypes {
		if !containsString(guide, typeName) {
			t.Errorf("InteractionTypeGuide should contain %q", typeName)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s[1:], substr) || s[:len(substr)] == substr)
}
