package formatter

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/buildkite/terminal-to-html/v3"

	"github.com/chubin/wttr.in/internal/assets"
	"github.com/chubin/wttr.in/internal/domain"
)

// HTMLFormatter converts ANSI weather output into a full HTML page using buildkite/terminal-to-html.
type HTMLFormatter struct {
	template *template.Template
	css      string
}

// NewHTMLFormatter loads the embedded template and CSS.
func NewHTMLFormatter() (*HTMLFormatter, error) {
	// Load the HTML template
	tmplData, err := assets.GetFile("share/templates/index_buildkite.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load HTML template: %w", err)
	}

	tmpl, err := template.New("weather-html").Parse(string(tmplData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	cssData, err := assets.GetFile("share/static/terminal.css")
	if err != nil {
		return nil, fmt.Errorf("failed to load terminal.css: %w", err)
	}
	css := string(cssData)

	return &HTMLFormatter{
		template: tmpl,
		css:      css,
	}, nil
}

// documentLanguage returns the html lang attribute and the document title
// for the requested language, so that browser tabs, screen readers and
// search engines see a localized page instead of a hardcoded English one.
func documentLanguage(query *domain.Query) (string, string) {
	lang := "en"

	if query != nil && query.Options != nil && query.Options.Lang != "" {
		lang = query.Options.Lang
	}

	switch lang {
	case "zh", "zh-cn":
		return "zh-CN", "天气预报"
	case "zh-tw":
		return "zh-TW", "天氣預報"
	}

	return lang, "Weather Report"
}

func (f *HTMLFormatter) Format(query *domain.Query, output *domain.RenderOutput) (*domain.FormatOutput, error) {
	if output.Content == nil || len(output.Content) == 0 {
		return nil, fmt.Errorf("no render output to convert to HTML")
	}

	lang, title := documentLanguage(query)

	// Convert ANSI → HTML using buildkite
	htmlBody := terminal.Render(output.Content)

	// Render the full document using the embedded template
	data := struct {
		Title string
		Lang  string
		CSS   template.CSS
		Query *domain.Query
		Body  template.HTML
	}{
		Title: title,
		Lang:  lang,
		CSS:   template.CSS(f.css),
		Query: query,
		Body:  template.HTML(htmlBody),
	}

	var buf bytes.Buffer
	if err := f.template.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return &domain.FormatOutput{
		Content:     buf.Bytes(),
		ContentType: "text/html; charset=utf-8",
	}, nil
}
