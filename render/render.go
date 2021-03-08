package render

import (
	"bytes"
	"encoding/json"
	"go/doc"
	"html/template"
	"strings"

	"github.com/fatih/structtag"
	"github.com/gobuffalo/plush"
	"github.com/markbates/inflect"
	"github.com/pacedotdev/oto/parser"
	"github.com/pkg/errors"
)

var defaultRuleset = inflect.NewDefaultRuleset()

// Render renders the template using the Definition.
func Render(template string, def parser.Definition, params map[string]interface{}) (string, error) {
	ctx := plush.NewContext()
	ctx.Set("camelize_down", camelizeDown)
	ctx.Set("snake_down", snakeDown)
	ctx.Set("def", def)
	ctx.Set("params", params)
	ctx.Set("json", toJSONHelper)
	ctx.Set("format_comment_line", formatCommentLine)
	ctx.Set("format_comment_text", formatCommentText)
	ctx.Set("format_comment_html", formatCommentHTML)
	ctx.Set("format_tags", formatTags)
	ctx.Set("strip_prefix", stripPrefix)
	ctx.Set("strip_suffix", stripSuffix)
	ctx.Set("has_prefix", strings.HasPrefix)
	ctx.Set("has_suffix", strings.HasSuffix)
	ctx.Set("to_lower", strings.ToLower)
	ctx.Set("to_upper", strings.ToUpper)
	s, err := plush.Render(string(template), ctx)
	if err != nil {
		return "", err
	}
	return s, nil
}

func toJSONHelper(v interface{}) (template.HTML, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

func formatCommentLine(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "", "", 2000)
	s = strings.TrimSpace(buf.String())
	return template.HTML(s)
}

func formatCommentText(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "// ", "", 80)
	return template.HTML(buf.String())
}

func formatCommentHTML(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToHTML(&buf, s, nil)
	return template.HTML(buf.String())
}

// formatTags formats a list of struct tag strings into one.
// Will return an error if any of the tag strings are invalid.
func formatTags(tags ...string) (template.HTML, error) {
	alltags := &structtag.Tags{}
	for _, tag := range tags {
		theseTags, err := structtag.Parse(tag)
		if err != nil {
			return "", errors.Wrapf(err, "parse tags: `%s`", tag)
		}
		for _, t := range theseTags.Tags() {
			alltags.Set(t)
		}
	}
	tagsStr := alltags.String()
	if tagsStr == "" {
		return "", nil
	}
	tagsStr = "`" + tagsStr + "`"
	return template.HTML(tagsStr), nil
}

func stripPrefix(s, prefix string) (string, error) {
	if !strings.HasPrefix(s, prefix) {
		return s, errors.Errorf("cannot strip prefix: %s from: %s", prefix, s)
	}
	return strings.TrimPrefix(s, prefix), nil
}

func stripSuffix(s, suffix string) (string, error) {
	if !strings.HasSuffix(s, suffix) {
		return s, errors.Errorf("cannot strip suffix: %s from: %s", suffix, s)
	}
	return strings.TrimSuffix(s, suffix), nil
}
