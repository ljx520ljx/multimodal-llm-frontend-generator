package prompt

// InteractionType defines the supported interaction types
type InteractionType string

const (
	// InteractionToggle represents show/hide toggle interactions
	InteractionToggle InteractionType = "toggle"

	// InteractionExpand represents expand/collapse interactions
	InteractionExpand InteractionType = "expand"

	// InteractionTabSwitch represents tab switching interactions
	InteractionTabSwitch InteractionType = "tab_switch"

	// InteractionFormState represents form state change interactions
	InteractionFormState InteractionType = "form_state"

	// InteractionHover represents hover state interactions
	InteractionHover InteractionType = "hover"

	// InteractionNavigation represents navigation/view switching interactions
	InteractionNavigation InteractionType = "navigation"
)

// InteractionPattern describes the code generation pattern for an interaction type
type InteractionPattern struct {
	Type         InteractionType
	Description  string
	StateType    string // "boolean" | "string" | "number" | "object"
	DefaultValue string
	EventHandler string // "onClick" | "onChange" | "onMouseEnter" | "onMouseLeave"
	CodeExample  string
}

// InteractionPatterns maps interaction types to their code generation patterns
var InteractionPatterns = map[InteractionType]InteractionPattern{
	InteractionToggle: {
		Type:         InteractionToggle,
		Description:  "元素显示/隐藏切换",
		StateType:    "boolean",
		DefaultValue: "false",
		EventHandler: "onClick",
		CodeExample:  `const [isOpen, setIsOpen] = useState(false);`,
	},
	InteractionExpand: {
		Type:         InteractionExpand,
		Description:  "区域展开/收起",
		StateType:    "boolean",
		DefaultValue: "false",
		EventHandler: "onClick",
		CodeExample:  `const [isExpanded, setIsExpanded] = useState(false);`,
	},
	InteractionTabSwitch: {
		Type:         InteractionTabSwitch,
		Description:  "Tab/标签页切换",
		StateType:    "string",
		DefaultValue: `"tab1"`,
		EventHandler: "onClick",
		CodeExample:  `const [activeTab, setActiveTab] = useState("tab1");`,
	},
	InteractionFormState: {
		Type:         InteractionFormState,
		Description:  "表单输入状态变化",
		StateType:    "object",
		DefaultValue: "{}",
		EventHandler: "onChange",
		CodeExample:  `const [formData, setFormData] = useState({ name: "", email: "" });`,
	},
	InteractionHover: {
		Type:         InteractionHover,
		Description:  "悬停样式变化",
		StateType:    "boolean",
		DefaultValue: "false",
		EventHandler: "onMouseEnter",
		CodeExample:  `// Prefer Tailwind hover: classes, e.g., hover:bg-blue-500`,
	},
	InteractionNavigation: {
		Type:         InteractionNavigation,
		Description:  "页面/视图跳转",
		StateType:    "string",
		DefaultValue: `"home"`,
		EventHandler: "onClick",
		CodeExample:  `const [currentView, setCurrentView] = useState("home");`,
	},
}

// GetAllInteractionTypes returns all supported interaction types
func GetAllInteractionTypes() []InteractionType {
	return []InteractionType{
		InteractionToggle,
		InteractionExpand,
		InteractionTabSwitch,
		InteractionFormState,
		InteractionHover,
		InteractionNavigation,
	}
}

// GetInteractionPattern returns the pattern for a given interaction type
func GetInteractionPattern(t InteractionType) (InteractionPattern, bool) {
	pattern, ok := InteractionPatterns[t]
	return pattern, ok
}

// InteractionTypeGuide returns a formatted guide for all interaction types
// This can be used in prompts to help LLM understand the types
func InteractionTypeGuide() string {
	return `| 类型 | 特征 | 代码模式 |
|------|------|----------|
| Toggle | 元素显示/隐藏切换 | useState<boolean> + onClick |
| Expand | 区域展开/收起 | useState<boolean> + 高度过渡 |
| TabSwitch | Tab/标签页切换 | useState<string> + onClick |
| FormState | 表单输入状态变化 | useState + onChange |
| Hover | 悬停样式变化 | hover: 类或 onMouseEnter |
| Navigation | 页面/视图跳转 | useState<string> |`
}
