package goquery

import "testing"

func BenchmarkValInput(b *testing.B) {
	var v string

	b.StopTimer()
	sel := formValDoc().Find(`input[name="text-value"]`)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		got, ok := sel.Val()
		if !ok {
			b.Fatal("expected ok=true")
		}
		v = got.(string)
	}
	if v != "hello" {
		b.Fatalf("want hello, got %q", v)
	}
}

func BenchmarkValTextarea(b *testing.B) {
	b.StopTimer()
	sel := formValDoc().Find(`textarea[name="ta-value"]`)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sel.Val()
	}
}

func BenchmarkValSelectSingle(b *testing.B) {
	var v string

	b.StopTimer()
	sel := formValDoc().Find(`select[name="select-single"]`)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		got, ok := sel.Val()
		if !ok {
			b.Fatal("expected ok=true")
		}
		v = got.(string)
	}
	if v != "b" {
		b.Fatalf("want b, got %q", v)
	}
}

func BenchmarkValSelectMultiple(b *testing.B) {
	b.StopTimer()
	sel := formValDoc().Find(`select[name="select-multi"]`)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sel.Val()
	}
}

func BenchmarkValCheckboxGroup(b *testing.B) {
	b.StopTimer()
	sel := formValDoc().Find(`input[name="c1"]`)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sel.Val()
	}
}
