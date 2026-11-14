package goquery

import (
	"reflect"
	"strings"
	"testing"
)

// formValFixture is the shared markup used by the Val matrix tests. Every
// control it declares is referenced by a selector in formValCases.
const formValFixture = `
<html><body><form id="f">
	<input name="text-default" type="text">
	<input name="text-value" type="text" value="hello">
	<input name="search" type="search" value="q">
	<input name="password" type="password" value="secret">
	<input name="hidden" type="hidden" value="h">
	<input name="number" type="number" value="42">
	<input name="date" type="date" value="2026-11-14">
	<input name="color" type="color" value="#0000ff">
	<input name="no-type" value="default-type">
	<input name="unknown-type" type="widget" value="u">
	<input name="file" type="file">
	<input name="submit" type="submit" value="Go">
	<input name="image" type="image" src="btn.png" value="img-val">
	<input name="button" type="button" value="Click">
	<input name="reset" type="reset">

	<textarea name="ta-empty"></textarea>
	<textarea name="ta-value">line1
line2 &amp; done</textarea>

	<select name="select-single">
		<option value="a">A</option>
		<option value="b" selected>B</option>
		<option value="c">C</option>
	</select>
	<select name="select-default-first">
		<option value="a">A</option>
		<option value="b">B</option>
	</select>
	<select name="select-text-value">
		<option>plain text</option>
		<option selected>chosen text</option>
	</select>
	<select name="select-disabled-first">
		<option value="x" disabled>X</option>
		<option value="y">Y</option>
	</select>
	<select name="select-all-disabled">
		<option value="x" disabled>X</option>
	</select>
	<select name="select-empty"></select>
	<select name="select-multi" multiple>
		<option value="a" selected>A</option>
		<option value="b">B</option>
		<option selected>C</option>
	</select>
	<select name="select-multi-none" multiple>
		<option value="a">A</option>
		<option value="b">B</option>
	</select>
	<select name="select-multi-optgroup" multiple>
		<optgroup label="g1"><option value="g1a" selected>G1A</option></optgroup>
		<optgroup label="g2"><option value="g2b" selected>G2B</option></optgroup>
	</select>

	<label><input type="radio" name="r1" value="one"></label>
	<label><input type="radio" name="r1" value="two" checked></label>
	<label><input type="radio" name="r1" value="three"></label>
	<input type="radio" name="r2" value="only">
	<input type="radio" name="r3" checked>
	<input type="radio" name="r4">

	<input type="checkbox" name="c1" value="one" checked>
	<input type="checkbox" name="c1" value="two">
	<input type="checkbox" name="c1" value="three" checked="checked">
	<input type="checkbox" name="c2" checked>
	<input type="checkbox" name="c3" value="only">

	<div id="not-a-control">plain</div>
	<button type="button" id="button-el">B</button>
</form></body></html>`

func formValDoc() *Document {
	d, err := NewDocumentFromReader(strings.NewReader(formValFixture))
	if err != nil {
		panic(err)
	}
	return d
}

// formValCase is one cell of the control-form x value-scenario matrix:
// selector picks the control (or group), want is the expected Val result and
// wantOk the expected ok flag.
type formValCase struct {
	name     string
	selector string
	want     any
	wantOk   bool
}

func formValCases() []formValCase {
	return []formValCase{
		// --- text-like single-value inputs, every shape returns a string ---
		{"text/empty", `input[name="text-default"]`, "", true},
		{"text/value", `input[name="text-value"]`, "hello", true},
		{"search/value", `input[name="search"]`, "q", true},
		{"password/value", `input[name="password"]`, "secret", true},
		{"hidden/value", `input[name="hidden"]`, "h", true},
		{"number/value-kept-as-string", `input[name="number"]`, "42", true},
		{"date/value-kept-as-string", `input[name="date"]`, "2026-11-14", true},
		{"color/value-kept-as-string", `input[name="color"]`, "#0000ff", true},
		{"input/no-type-defaults-to-text", `input[name="no-type"]`, "default-type", true},
		{"input/unknown-type-treated-as-text", `input[name="unknown-type"]`, "u", true},
		{"file/empty-string", `input[name="file"]`, "", true},
		{"submit/value", `input[name="submit"]`, "Go", true},
		{"image/value", `input[name="image"]`, "img-val", true},
		{"button/value", `input[name="button"]`, "Click", true},
		{"reset/empty-string", `input[name="reset"]`, "", true},

		// --- textarea: string, raw rendered contents, entities decoded ---
		{"textarea/empty", `textarea[name="ta-empty"]`, "", true},
		{"textarea/multiline-and-entity", `textarea[name="ta-value"]`, "line1\nline2 & done", true},

		// --- single select: string ---
		{"select/single-selected", `select[name="select-single"]`, "b", true},
		{"select/defaults-to-first-option", `select[name="select-default-first"]`, "a", true},
		{"select/option-text-as-value", `select[name="select-text-value"]`, "chosen text", true},
		{"select/first-enabled-when-first-disabled", `select[name="select-disabled-first"]`, "y", true},
		{"select/all-disabled-empty-string", `select[name="select-all-disabled"]`, "", true},
		{"select/no-options-empty-string", `select[name="select-empty"]`, "", true},

		// --- multiple select: []string, never nil ---
		{"select/multiple-selected-in-order", `select[name="select-multi"]`, []string{"a", "C"}, true},
		{"select/multiple-none-empty-slice", `select[name="select-multi-none"]`, []string{}, true},
		{"select/multiple-optgroup-options", `select[name="select-multi-optgroup"]`, []string{"g1a", "g2b"}, true},

		// --- radio groups: string when checked, nil when not ---
		{"radio/checked-member-of-group", `input[name="r1"]`, "two", true},
		{"radio/unchecked-group", `input[name="r2"]`, nil, true},
		{"radio/checked-without-value-defaults-on", `input[name="r3"]`, "on", true},
		{"radio/unchecked-without-value", `input[name="r4"]`, nil, true},

		// --- checkbox groups: []string, never nil ---
		{"checkbox/checked-members-in-order", `input[name="c1"]`, []string{"one", "three"}, true},
		{"checkbox/checked-without-value-defaults-on", `input[name="c2"]`, []string{"on"}, true},
		{"checkbox/none-checked-empty-slice", `input[name="c3"]`, []string{}, true},

		// --- non-form elements and empty selections: ok=false, no regression ---
		{"non-form/div", "#not-a-control", nil, false},
		{"non-form/button-element", "#button-el", nil, false},
		{"empty/selection", "aside", nil, false},
	}
}

func TestValMatrix(t *testing.T) {
	doc := formValDoc()
	for _, tc := range formValCases() {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := doc.Find(tc.selector).Val()
			if ok != tc.wantOk {
				t.Fatalf("%s: ok = %v, want %v (got value %v)", tc.name, ok, tc.wantOk, got)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%s: value = %#v (%T), want %#v (%T)", tc.name, got, got, tc.want, tc.want)
			}
		})
	}
}

// TestValReturnType documents the value shape per control form: strings for
// single-value controls, []string for multi-value controls, nil for
// unchecked radio groups and non-form nodes.
func TestValReturnTypes(t *testing.T) {
	doc := formValDoc()

	assertType := func(name, selector string, want reflect.Type) {
		t.Helper()
		got, ok := doc.Find(selector).Val()
		if !ok {
			t.Fatalf("%s: expected ok=true", name)
		}
		if got != nil && reflect.TypeOf(got) != want {
			t.Errorf("%s: got type %T, want %s", name, got, want)
		}
	}

	assertType("text", `input[name="text-value"]`, reflect.TypeOf(""))
	assertType("textarea", `textarea[name="ta-value"]`, reflect.TypeOf(""))
	assertType("single-select", `select[name="select-single"]`, reflect.TypeOf(""))
	assertType("multiple-select", `select[name="select-multi"]`, reflect.TypeOf([]string{}))
	assertType("checkbox", `input[name="c1"]`, reflect.TypeOf([]string{}))
}

// TestValRadioFirstNode ensures the checked member is found wherever it sits
// in the Selection, and that a Selection reduced to a single unchecked radio
// still reports nil rather than a value.
func TestValRadioGroupPositions(t *testing.T) {
	doc := formValDoc()

	// The first node is unchecked but the second is checked.
	if got, ok := doc.Find(`input[name="r1"]`).Val(); !ok || got != "two" {
		t.Fatalf("radio group: got %v, ok=%v, want %q, ok=true", got, ok, "two")
	}

	// A single unchecked member must not pick up a sibling's value.
	if got, ok := doc.Find(`input[name="r1"]`).First().Val(); !ok || got != nil {
		t.Fatalf("single unchecked radio: got %v, ok=%v, want nil, ok=true", got, ok)
	}
}

// TestValCheckboxSingleSelection covers a Selection holding exactly one
// checkbox node, both checked and unchecked.
func TestValCheckboxSingleSelection(t *testing.T) {
	doc := formValDoc()

	got, ok := doc.Find(`input[name="c1"]`).Eq(0).Val()
	if !ok || !reflect.DeepEqual(got, []string{"one"}) {
		t.Fatalf("single checked checkbox: got %#v, ok=%v", got, ok)
	}

	got, ok = doc.Find(`input[name="c1"]`).Eq(1).Val()
	if !ok || !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("single unchecked checkbox: got %#v, ok=%v", got, ok)
	}
}

// TestValGroupIgnoresNonMatchingNodes verifies the group helpers ignore
// non-matching nodes in the Selection: a checkbox Selection keeps only the
// checked checkbox values even when other elements sit between them.
func TestValGroupIgnoresNonMatchingNodes(t *testing.T) {
	doc := formValDoc()

	// Every form field in one Selection, first node being a text input: the
	// first node decides the shape, so this reads as a plain text value.
	all := doc.Find("#f input, #f select, #f textarea")
	if got, ok := all.Val(); !ok || got != "" {
		t.Fatalf("first node is a text input: got %#v, ok=%v, want %q, ok=true", got, ok, "")
	}

	// The checkbox group plus an unrelated div in the same Selection: the
	// div is skipped and the checked boxes survive in document order.
	group := doc.Find(`input[name="c1"], #not-a-control, input[name="c2"]`)
	got, ok := group.Val()
	want := []string{"one", "three", "on"}
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("checkbox group with mixed nodes: got %#v, ok=%v, want %#v", got, ok, want)
	}
}

// TestValDoesNotMutateDocument ensures reading values leaves the parsed
// document (attributes and text) untouched.
func TestValDoesNotMutateDocument(t *testing.T) {
	doc := formValDoc()

	before, _ := doc.Find(`select[name="select-single"]`).Html()
	doc.Find(`select[name="select-single"]`).Val()
	doc.Find(`input[name="text-value"]`).Val()
	doc.Find(`textarea[name="ta-value"]`).Val()
	doc.Find(`input[name="r1"]`).Val()
	doc.Find(`input[name="c1"]`).Val()
	after, _ := doc.Find(`select[name="select-single"]`).Html()
	if before != after {
		t.Fatalf("Val mutated the document:\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestValAttrStillWorks guards the existing attribute-reading product against
// regression while the new capability lives in the same package.
func TestValAttrStillWorks(t *testing.T) {
	doc := formValDoc()

	if v, ok := doc.Find(`input[name="text-value"]`).Attr("value"); !ok || v != "hello" {
		t.Fatalf("Attr after Val: got %q, ok=%v, want %q", v, ok, "hello")
	}
	if v, ok := doc.Find(`select[name="select-single"] option`).Eq(1).Attr("selected"); !ok || v != "" {
		t.Fatalf("boolean attribute readback changed: got %q, ok=%v", v, ok)
	}
}
