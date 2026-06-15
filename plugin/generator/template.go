package generator

import (
	"fmt"

	"github.com/the-protobuf-project/grpc-mcp-gateway/plugin/generator/templates"
)

// Supported languages for code generation.
const (
	LangGo     = "go"
	LangPython = "python"
	LangRust   = "rust"
	LangCpp    = "cpp"
)

// codeTemplates maps language name → embedded template content.
var codeTemplates = map[string]string{
	LangGo:     mustReadTemplate("go.tpl"),
	LangPython: mustReadTemplate("python.tpl"),
	LangRust:   mustReadTemplate("rust.tpl"),
	LangCpp:    mustReadTemplate("cpp/mcp.h.tpl"),
}

// GetTemplate returns the code template for the given language.
func GetTemplate(lang string) (string, error) {
	tpl, ok := codeTemplates[lang]
	if !ok {
		return "", fmt.Errorf("unsupported language %q (supported: go, python, rust, cpp)", lang)
	}
	return tpl, nil
}

func mustReadTemplate(name string) string {
	b, err := templates.FS.ReadFile(name)
	if err != nil {
		panic("generator: embedded template " + name + " not found: " + err.Error())
	}
	return string(b)
}
