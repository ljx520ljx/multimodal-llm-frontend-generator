package prompt

import (
	"strings"
	"testing"
)

func TestBuildSystemPromptV2(t *testing.T) {
	tests := []struct {
		framework string
		contains  []string
	}{
		{
			framework: "react",
			contains:  []string{"React", "核心能力", "技术约束", "输出格式", "<thinking>"},
		},
		{
			framework: "vue",
			contains:  []string{"Vue", "核心能力", "技术约束"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			result := BuildSystemPromptV2(tt.framework)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("BuildSystemPromptV2(%q) should contain %q", tt.framework, substr)
				}
			}
		})
	}
}

func TestBuildSingleImagePromptV2(t *testing.T) {
	result := BuildSingleImagePromptV2("react")

	expectedContents := []string{
		"Step 1: 布局识别",
		"Step 2: 组件识别",
		"Step 3: 样式分析",
		"Step 4: 代码生成",
		"React",
	}

	for _, content := range expectedContents {
		if !strings.Contains(result, content) {
			t.Errorf("BuildSingleImagePromptV2 should contain %q", content)
		}
	}
}

func TestBuildMultiImagePromptV2(t *testing.T) {
	tests := []struct {
		imageCount int
		framework  string
	}{
		{2, "react"},
		{3, "react"},
		{5, "vue"},
	}

	for _, tt := range tests {
		result := BuildMultiImagePromptV2(tt.imageCount, tt.framework)

		// Should contain the image count
		if !strings.Contains(result, "张 UI 设计稿") {
			t.Errorf("BuildMultiImagePromptV2(%d, %q) should mention images", tt.imageCount, tt.framework)
		}

		// Should contain analysis steps
		expectedSteps := []string{
			"Step 1: 布局识别",
			"Step 2: 组件识别",
			"Step 3: 差异检测",
			"Step 4: 交互推理",
			"Step 5: 代码生成",
		}

		for _, step := range expectedSteps {
			if !strings.Contains(result, step) {
				t.Errorf("BuildMultiImagePromptV2 should contain %q", step)
			}
		}

		// Should contain interaction types
		interactionTypes := []string{"Toggle", "Expand", "TabSwitch", "FormState", "Hover", "Navigation"}
		for _, iType := range interactionTypes {
			if !strings.Contains(result, iType) {
				t.Errorf("BuildMultiImagePromptV2 should contain interaction type %q", iType)
			}
		}
	}
}

func TestBuildChatModifyPromptV2(t *testing.T) {
	code := "export default function App() { return <div>Hello</div> }"
	message := "把文字改成红色"

	result := BuildChatModifyPromptV2(code, message)

	if !strings.Contains(result, code) {
		t.Error("BuildChatModifyPromptV2 should contain the original code")
	}
	if !strings.Contains(result, message) {
		t.Error("BuildChatModifyPromptV2 should contain the user message")
	}
	if !strings.Contains(result, "<thinking>") {
		t.Error("BuildChatModifyPromptV2 should mention <thinking> format")
	}
}

func TestDiffAnalysisPromptV2Content(t *testing.T) {
	prompt := DiffAnalysisPromptV2

	expectedContents := []string{
		"双图差异分析",
		"Step 1:",
		"Step 2:",
		"Step 3:",
		"Step 4:",
		"交互类型",
		"状态变量",
	}

	for _, content := range expectedContents {
		if !strings.Contains(prompt, content) {
			t.Errorf("DiffAnalysisPromptV2 should contain %q", content)
		}
	}
}
