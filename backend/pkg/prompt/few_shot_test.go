package prompt

import (
	"strings"
	"testing"
)

func TestFewShotExamples(t *testing.T) {
	// Should have at least 2 examples
	if len(FewShotExamples) < 2 {
		t.Errorf("Expected at least 2 few-shot examples, got %d", len(FewShotExamples))
	}

	// Each example should have all fields populated
	for _, example := range FewShotExamples {
		if example.Name == "" {
			t.Error("Example Name should not be empty")
		}
		if example.Description == "" {
			t.Errorf("Example %q Description should not be empty", example.Name)
		}
		if example.Input == "" {
			t.Errorf("Example %q Input should not be empty", example.Name)
		}
		if example.Thinking == "" {
			t.Errorf("Example %q Thinking should not be empty", example.Name)
		}
		if example.Code == "" {
			t.Errorf("Example %q Code should not be empty", example.Name)
		}
	}
}

func TestGetFewShotExample(t *testing.T) {
	tests := []struct {
		name      string
		wantFound bool
	}{
		{"toggle_modal", true},
		{"tab_switch", true},
		{"nonexistent", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			example, found := GetFewShotExample(tt.name)

			if found != tt.wantFound {
				t.Errorf("GetFewShotExample(%q) found = %v, want %v", tt.name, found, tt.wantFound)
			}

			if found && example.Name != tt.name {
				t.Errorf("GetFewShotExample(%q) returned example with Name = %q", tt.name, example.Name)
			}

			if !found && example.Name != "" {
				t.Errorf("GetFewShotExample(%q) should return empty example when not found", tt.name)
			}
		})
	}
}

func TestFormatFewShotForPrompt(t *testing.T) {
	example := FewShotExample{
		Name:        "test_example",
		Description: "Test description",
		Input:       "Test input",
		Thinking:    "Test thinking",
		Code:        "Test code",
	}

	result := FormatFewShotForPrompt(example)

	// Should contain all parts
	expectedParts := []string{
		"### 示例",
		"**输入图片描述**: Test input",
		"**分析过程**:",
		"<thinking>",
		"Test thinking",
		"</thinking>",
		"**输出代码**:",
		"```jsx",
		"Test code",
		"```",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("FormatFewShotForPrompt should contain %q", part)
		}
	}
}

func TestFormatFewShotForPrompt_WithRealExample(t *testing.T) {
	example, found := GetFewShotExample("toggle_modal")
	if !found {
		t.Fatal("toggle_modal example should exist")
	}

	result := FormatFewShotForPrompt(example)

	// Should contain analysis steps from the example
	if !strings.Contains(result, "Step 1:") {
		t.Error("Formatted prompt should contain Step 1")
	}
	if !strings.Contains(result, "Step 4:") {
		t.Error("Formatted prompt should contain Step 4")
	}
	if !strings.Contains(result, "useState") {
		t.Error("Formatted prompt should contain useState in code")
	}
}

func TestGetDefaultFewShotPrompt_Disabled(t *testing.T) {
	result := GetDefaultFewShotPrompt(false)

	if result != "" {
		t.Errorf("GetDefaultFewShotPrompt(false) should return empty string, got %q", result)
	}
}

func TestGetDefaultFewShotPrompt_Enabled(t *testing.T) {
	result := GetDefaultFewShotPrompt(true)

	// Should not be empty
	if result == "" {
		t.Error("GetDefaultFewShotPrompt(true) should not return empty string")
	}

	// Should contain example formatting
	expectedParts := []string{
		"### 示例",
		"<thinking>",
		"</thinking>",
		"```jsx",
		"现在请分析用户提供的图片",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("GetDefaultFewShotPrompt(true) should contain %q", part)
		}
	}

	// Should use the first example (toggle_modal)
	if !strings.Contains(result, "isModalOpen") {
		t.Error("GetDefaultFewShotPrompt should use toggle_modal example")
	}
}

func TestGetFrameworkDisplayName_Default(t *testing.T) {
	// Test default case (unknown framework)
	result := GetFrameworkDisplayName("unknown")
	if result != ReactDisplayName {
		t.Errorf("GetFrameworkDisplayName(\"unknown\") = %q, want %q", result, ReactDisplayName)
	}

	// Test empty string
	result = GetFrameworkDisplayName("")
	if result != ReactDisplayName {
		t.Errorf("GetFrameworkDisplayName(\"\") = %q, want %q", result, ReactDisplayName)
	}
}

func TestFewShotExamples_QualityChecks(t *testing.T) {
	for _, example := range FewShotExamples {
		t.Run(example.Name, func(t *testing.T) {
			// Thinking should contain analysis steps
			if !strings.Contains(example.Thinking, "Step 1:") {
				t.Errorf("Example %q Thinking should contain Step 1", example.Name)
			}
			if !strings.Contains(example.Thinking, "Step 4:") {
				t.Errorf("Example %q Thinking should contain Step 4 (交互推理)", example.Name)
			}

			// Code should be valid React code
			if !strings.Contains(example.Code, "import") {
				t.Errorf("Example %q Code should contain import statement", example.Name)
			}
			if !strings.Contains(example.Code, "export default") {
				t.Errorf("Example %q Code should contain export default", example.Name)
			}
			if !strings.Contains(example.Code, "useState") {
				t.Errorf("Example %q Code should contain useState for state management", example.Name)
			}
		})
	}
}
