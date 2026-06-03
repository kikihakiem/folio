// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"strings"
	"testing"

	"github.com/kikihakiem/folio/document"
	"github.com/kikihakiem/folio/font"
	"github.com/kikihakiem/folio/forms"
	"github.com/kikihakiem/folio/reader"
)

func buildPDFWithForm(t *testing.T) []byte {
	t.Helper()
	doc := document.NewDocument(document.PageSizeLetter)
	doc.AddPage()

	form := forms.NewAcroForm()
	form.Add(forms.NewTextField("name", [4]float64{72, 700, 300, 720}, 0))
	form.Add(forms.NewCheckbox("agree", [4]float64{72, 670, 92, 690}, 0, true))
	doc.SetAcroForm(form)

	var buf strings.Builder
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	return []byte(buf.String())
}

func TestFlattenFormsBasic(t *testing.T) {
	pdfBytes := buildPDFWithForm(t)

	r, err := reader.Parse(pdfBytes)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	m, err := reader.Merge(r)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	err = m.FlattenForms()
	if err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	var out strings.Builder
	if _, err := m.WriteTo(&out); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected non-empty output")
	}

	// Parse the result — should be valid PDF.
	r2, err := reader.Parse([]byte(out.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	_ = r2
}

func TestFlattenFormsNoForms(t *testing.T) {
	doc := document.NewDocument(document.PageSizeLetter)
	doc.AddPage()

	var buf strings.Builder
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	r, err := reader.Parse([]byte(buf.String()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	m, err := reader.Merge(r)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	err = m.FlattenForms()
	if err != nil {
		t.Fatalf("FlattenForms on formless PDF: %v", err)
	}
}

func TestExtractPageImportBasic(t *testing.T) {
	doc := document.NewDocument(document.PageSizeLetter)
	p := doc.AddPage()
	p.AddText("Hello World", font.Helvetica, 12, 100, 400)

	var buf strings.Builder
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	r, err := reader.Parse([]byte(buf.String()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	pi, err := reader.ExtractPageImport(r, 0)
	if err != nil {
		t.Fatalf("ExtractPageImport: %v", err)
	}

	if pi.Width <= 0 || pi.Height <= 0 {
		t.Errorf("expected positive dimensions, got %gx%g", pi.Width, pi.Height)
	}
	if pi.Resources == nil {
		t.Error("expected non-nil resources")
	}
}

func TestExtractPageImportOutOfRange(t *testing.T) {
	doc := document.NewDocument(document.PageSizeLetter)
	doc.AddPage()

	var buf strings.Builder
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	r, err := reader.Parse([]byte(buf.String()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, err = reader.ExtractPageImport(r, 99)
	if err == nil {
		t.Error("expected error for out-of-range page index")
	}
}
