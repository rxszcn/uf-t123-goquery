package goquery

import (
	"strings"

	"golang.org/x/net/html"
)

// Val reads the current value of the first form control in the Selection,
// using only the information present in the parsed HTML document (there is no
// live DOM, so the value is what the markup declares: the value attribute,
// selected/checked flags and the text contents of textarea and option
// elements).
//
// The ok result reports whether the first node of the Selection is a form
// control that Val can read. It is false when the Selection is empty or its
// first node is not an input, textarea or select element (for example a div,
// a button element, or a text node). In that case value is nil. Val never
// panics on those inputs, and it does not change any existing behaviour:
// reading attributes through Attr and reading non-form elements keep working
// exactly as before.
//
// The concrete type of value depends on the form control:
//
//   - text-like inputs (the default type, as well as text, search, url,
//     tel, email, password, hidden, number, range, color, date, time and the
//     other date/time types), file, submit, reset, button and image inputs:
//     a string, taken from the value attribute, which defaults to "".
//   - textarea: a string holding the element's rendered text contents
//     (entity references are decoded by the HTML parser), which defaults to
//     "".
//   - single select (a select without the multiple attribute): a string with
//     the value of the selected option. An option's value is its value
//     attribute, or its text contents when that attribute is absent. When no
//     option carries the selected attribute, the first non-disabled option is
//     selected, mirroring what a browser submits; a select with no option
//     yields "".
//   - radio inputs: a string with the value of the checked input. The
//     Selection may hold the whole radio group (several inputs sharing a
//     name); the checked member's value is returned regardless of its
//     position. A checked input without a value attribute yields "on", the
//     browser default. When nothing is checked, value is nil.
//   - checkbox inputs and multiple select: a []string, always non-nil when ok
//     is true, containing every checked input's value (checkbox) or every
//     selected option's value (multiple select), in document order. Each
//     item defaults to "on" (checkbox) or the option's text contents (select)
//     the same way as the single-value cases. An empty slice means nothing is
//     checked or selected.
//
// For radio and checkbox, the first element of the Selection determines the
// group semantics, and every matching input of the same type in the
// Selection contributes to the result while other nodes are ignored. For
// every other control, only the first element of the Selection is examined,
// mirroring jQuery's Val(). An input with an unknown type is treated as a
// text-like input, as required by the HTML specification.
func (s *Selection) Val() (value any, ok bool) {
	if len(s.Nodes) == 0 {
		return nil, false
	}
	n := s.Nodes[0]
	if n == nil || n.Type != html.ElementNode {
		return nil, false
	}

	switch n.Data {
	case "input":
		return inputVal(s.Nodes, n)
	case "textarea":
		return textareaVal(n), true
	case "select":
		return selectVal(n), true
	default:
		return nil, false
	}
}

// inputVal dispatches an input element to the checkbox/radio group logic or
// to the scalar value-attribute logic. The group logic scans the whole
// Selection so the first node need not be the checked one.
func inputVal(nodes []*html.Node, first *html.Node) (any, bool) {
	typ := strings.ToLower(inputType(first))
	switch typ {
	case "checkbox":
		return checkedInputValues(nodes, "checkbox"), true
	case "radio":
		for _, node := range nodes {
			if isInputOfType(node, "radio") && getAttributePtr("checked", node) != nil {
				return checkedInputValue(node), true
			}
		}
		return nil, true
	default:
		return inputValue(first), true
	}
}

// inputValue returns the submitted value of a text-like input element: the
// value attribute, or "" when it is absent.
func inputValue(n *html.Node) string {
	if attr := getAttributePtr("value", n); attr != nil {
		return attr.Val
	}
	return ""
}

// checkedInputValue returns the submitted value of a checked radio or
// checkbox input: the value attribute, or the browser default "on" when it is
// absent.
func checkedInputValue(n *html.Node) string {
	if attr := getAttributePtr("value", n); attr != nil {
		return attr.Val
	}
	return "on"
}

// checkedInputValues collects the value of every checked input of the given
// type across the Selection, in document order. Nodes that are not matching
// input elements are skipped, so a Selection scoped to a whole form still
// yields the group's values. Checked inputs without a value attribute
// contribute the browser default "on". The returned slice is never nil.
func checkedInputValues(nodes []*html.Node, typ string) []string {
	values := []string{}
	for _, n := range nodes {
		if !isInputOfType(n, typ) {
			continue
		}
		if getAttributePtr("checked", n) == nil {
			continue
		}
		values = append(values, checkedInputValue(n))
	}
	return values
}

// isInputOfType reports whether n is an input element whose type attribute
// matches the given (lower-case) type.
func isInputOfType(n *html.Node, typ string) bool {
	if n == nil || n.Type != html.ElementNode || n.Data != "input" {
		return false
	}
	return strings.ToLower(inputType(n)) == typ
}

// inputType returns the lower-cased type attribute of an input element, or
// "text" when the attribute is absent, per the HTML specification default.
func inputType(n *html.Node) string {
	if attr := getAttributePtr("type", n); attr != nil {
		return strings.ToLower(attr.Val)
	}
	return "text"
}

// textareaVal returns the current value of a textarea element, which is its
// rendered text contents in the same shape as Selection.Text.
func textareaVal(n *html.Node) string {
	return newSingleSelection(n, nil).Text()
}

// selectVal reads the value of a select element. A multiple select returns
// every selected option's value as a []string (never nil), while a single
// select returns the selected option's value as a string, defaulting to the
// first non-disabled option when nothing carries the selected attribute.
func selectVal(n *html.Node) any {
	var selected []string
	firstEnabled := ""
	haveFirstEnabled := false

	for _, option := range findOptions(n) {
		disabled := getAttributePtr("disabled", option) != nil
		if !haveFirstEnabled && !disabled {
			firstEnabled = optionValue(option)
			haveFirstEnabled = true
		}
		if getAttributePtr("selected", option) != nil {
			selected = append(selected, optionValue(option))
		}
	}

	if getAttributePtr("multiple", n) != nil {
		if selected == nil {
			return []string{}
		}
		return selected
	}

	if len(selected) > 0 {
		return selected[len(selected)-1]
	}
	if haveFirstEnabled {
		return firstEnabled
	}
	return ""
}

// findOptions returns every option element nested under n, in document order,
// including options wrapped in optgroup elements.
func findOptions(n *html.Node) []*html.Node {
	var options []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "option" {
				options = append(options, c)
				continue
			}
			walk(c)
		}
	}
	walk(n)
	return options
}

// optionValue returns the submitted value of an option element: its value
// attribute, or its rendered text contents when the attribute is absent.
func optionValue(n *html.Node) string {
	if attr := getAttributePtr("value", n); attr != nil {
		return attr.Val
	}
	return newSingleSelection(n, nil).Text()
}
