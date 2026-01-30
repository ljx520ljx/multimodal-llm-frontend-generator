package prompt

// FewShotExample represents a single few-shot example for prompt engineering
type FewShotExample struct {
	Name        string // Example name for reference
	Description string // Brief description of the example
	Input       string // Description of input images
	Thinking    string // Expected thinking/analysis process
	Code        string // Expected output code
}

// FewShotExamples contains high-quality examples for multi-image analysis
var FewShotExamples = []FewShotExample{
	{
		Name:        "toggle_modal",
		Description: "点击按钮显示/隐藏模态框",
		Input:       "图1: 页面有一个「查看详情」按钮；图2: 页面中央显示一个模态框，有关闭按钮",
		Thinking: `**Step 1: 布局识别**
图1: 简单页面布局，中央有一个操作按钮
图2: 同样的页面，但覆盖了一个半透明遮罩和居中的模态框

**Step 2: 组件识别**
- MainPage: 主页面容器
- TriggerButton: 触发按钮「查看详情」
- Modal: 模态框组件
- Overlay: 半透明遮罩层
- CloseButton: 关闭按钮

**Step 3: 差异检测**
- 图2 新增: Modal 组件（包含标题、内容、关闭按钮）
- 图2 新增: Overlay 半透明遮罩
- 按钮样式无变化

**Step 4: 交互推理**
- 交互类型: Toggle
- 触发元素: TriggerButton (onClick)
- 关闭触发: CloseButton 或 Overlay (onClick)
- 状态变量: isModalOpen (boolean)`,
		Code: `import { useState } from "react";

export default function App() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center">
      <button
        onClick={() => setIsModalOpen(true)}
        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        查看详情
      </button>

      {isModalOpen && (
        <>
          <div
            className="fixed inset-0 bg-black bg-opacity-50"
            onClick={() => setIsModalOpen(false)}
          />
          <div className="fixed inset-0 flex items-center justify-center pointer-events-none">
            <div className="bg-white rounded-lg p-6 shadow-xl pointer-events-auto">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold">详情</h2>
                <button
                  onClick={() => setIsModalOpen(false)}
                  className="text-gray-500 hover:text-gray-700"
                >
                  ✕
                </button>
              </div>
              <p className="text-gray-600">这是模态框的内容</p>
            </div>
          </div>
        </>
      )}
    </div>
  );
}`,
	},
	{
		Name:        "tab_switch",
		Description: "Tab 标签页切换",
		Input:       "图1: 有3个Tab，第1个高亮，下方显示对应内容A；图2: 第2个Tab高亮，下方显示内容B",
		Thinking: `**Step 1: 布局识别**
图1: 顶部 Tab 栏 + 内容区域，Tab1 激活
图2: 相同布局，Tab2 激活

**Step 2: 组件识别**
- TabBar: Tab 栏容器
- TabItem: 单个 Tab 项（3个）
- ContentArea: 内容展示区域

**Step 3: 差异检测**
- Tab1: 图1高亮(蓝色) → 图2普通(灰色)
- Tab2: 图1普通(灰色) → 图2高亮(蓝色)
- 内容区: 图1显示"内容A" → 图2显示"内容B"

**Step 4: 交互推理**
- 交互类型: TabSwitch
- 触发元素: TabItem (onClick)
- 状态变量: activeTab (string: "tab1" | "tab2" | "tab3")`,
		Code: `import { useState } from "react";

export default function App() {
  const [activeTab, setActiveTab] = useState("tab1");

  const tabs = [
    { id: "tab1", label: "标签一", content: "这是标签一的内容" },
    { id: "tab2", label: "标签二", content: "这是标签二的内容" },
    { id: "tab3", label: "标签三", content: "这是标签三的内容" },
  ];

  return (
    <div className="p-4">
      <div className="flex border-b">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={` + "`" + `px-4 py-2 font-medium ${
              activeTab === tab.id
                ? "text-blue-500 border-b-2 border-blue-500"
                : "text-gray-500 hover:text-gray-700"
            }` + "`" + `}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="p-4">
        {tabs.find((tab) => tab.id === activeTab)?.content}
      </div>
    </div>
  );
}`,
	},
}

// GetFewShotExample returns a specific few-shot example by name
func GetFewShotExample(name string) (FewShotExample, bool) {
	for _, example := range FewShotExamples {
		if example.Name == name {
			return example, true
		}
	}
	return FewShotExample{}, false
}

// FormatFewShotForPrompt formats a few-shot example for inclusion in a prompt
func FormatFewShotForPrompt(example FewShotExample) string {
	return `### 示例

**输入图片描述**: ` + example.Input + `

**分析过程**:
<thinking>
` + example.Thinking + `
</thinking>

**输出代码**:
` + "```" + `jsx
` + example.Code + `
` + "```" + `
`
}

// GetDefaultFewShotPrompt returns the default few-shot section for prompts
// Returns empty string if few-shot is disabled
func GetDefaultFewShotPrompt(enabled bool) string {
	if !enabled {
		return ""
	}

	// Use the first example (toggle_modal) as it demonstrates core concepts
	example := FewShotExamples[0]
	return `

---

` + FormatFewShotForPrompt(example) + `
---

现在请分析用户提供的图片：`
}
