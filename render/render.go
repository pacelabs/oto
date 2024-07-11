package render

import (
	"bytes"
	"encoding/json"
	"go/doc"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatih/structtag"
	"github.com/gobuffalo/plush"
	"github.com/markbates/inflect"
	"github.com/pkg/errors"

	"github.com/meitner-se/oto/parser"
)

var defaultRuleset = inflect.NewDefaultRuleset()

// Render renders the template using the Definition.
func Render(template string, def parser.Definition, params map[string]interface{}) (string, error) {
	ctx := plush.NewContext()
	ctx.Set("camelize_down", camelizeDown)
	ctx.Set("camelize_up", camelizeUp)
	ctx.Set("snake_down", snakeDown)
	ctx.Set("def", def)
	ctx.Set("params", params)
	ctx.Set("json", toJSONHelper)
	ctx.Set("json_inline", toJSONInlineHelper)
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
	ctx.Set("is_number", regexp.MustCompile("^\\d+$").MatchString)
	ctx.Set("is_number_prefix", func(str string) bool { return len(str) > 0 && str[0] >= '0' && str[0] <= '9' })
	s, err := plush.Render(string(template), ctx)
	if err != nil {
		return "", err
	}
	return s, nil
}

func toJSONHelper(v interface{}) (template.HTML, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

func toJSONInlineHelper(v interface{}) (template.HTML, error) {
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

// isNumberPrefix checks if the prefix of the given string is a number
func isNumberPrefix(s string, length int) bool {
	if len(s) < length {
		return false
	}

	prefix := s[:length]

	// Check if the prefix consists entirely of digits
	for _, r := range prefix {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Optionally, try to parse the prefix as an integer
	_, err := strconv.Atoi(prefix)
	return err == nil
}
