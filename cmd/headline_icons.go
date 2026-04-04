package cmd

var headlineIcons = map[string]string{
	"profile":         "📦",
	"status":          "📦",
	"list":            "📦",
	"scan":            "🔍",
	"drift":           "🔍",
	"diff":            "🔍",
	"restore":         "🚀",
	"publish":         "🚀",
	"clone":           "🧬",
	"doctor":          "🩺",
	"import":          "📥",
	"brewfile-import": "📥",
	"delete":          "🗑",
	"update":          "🔄",
	"brewfile-export": "📦",
}

func headlineIcon(key string) string {
	if icon, ok := headlineIcons[key]; ok {
		return icon
	}
	return "📦"
}
