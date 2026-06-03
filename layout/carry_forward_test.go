// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package layout

import (
	"fmt"
	"testing"

	"github.com/kikihakiem/folio/font"
)

// buildCarryTable builds a header + 12 body rows table with a running-balance
// column, optionally configured with carried/brought-forward templates.
func buildCarryTable(withCarry bool) *Table {
	tbl := NewTable()
	tbl.SetColumnWidths([]float64{120, 80})

	h := tbl.AddHeaderRow()
	h.AddCell("Date", font.HelveticaBold, 10)
	h.AddCell("Balance", font.HelveticaBold, 10)

	for i := 0; i < 12; i++ {
		r := tbl.AddRow()
		r.AddCell(fmt.Sprintf("row %d", i), font.Helvetica, 10)
		// Balance cell as a Paragraph element (mirrors the HTML converter, which
		// builds <td> content as elements) so the carry value extraction must
		// read the element text, not Cell.text.
		r.AddCellElement(NewParagraph(fmt.Sprintf("%d", (i+1)*100), font.Helvetica, 10))
	}

	if withCarry {
		// Value cell as a Paragraph element (mirrors the HTML converter) so the
		// clone-and-fill must preserve the element's font via Paragraph.WithText.
		carried := NewRow()
		carried.AddCell("Carried forward", font.Helvetica, 9)
		carried.AddCellElement(NewParagraph("0", font.Helvetica, 9))

		brought := NewRow()
		brought.AddCell("Brought forward", font.Helvetica, 9)
		brought.AddCellElement(NewParagraph("0", font.Helvetica, 9))

		tbl.SetCarriedRow(carried, 1, 1) // value cell index 1, balance is column 1
		tbl.SetBroughtRow(brought, 1, 1)
	}

	return tbl
}

func firstBodyRow(t *Table) *Row {
	for _, r := range t.rows {
		if !r.isHeader && !r.isFooter {
			return r
		}
	}

	return nil
}

// TestTableCarryForward verifies the continuation gets a brought-forward row,
// prepended after the header rows and filled with the boundary running balance.
func TestTableCarryForward(t *testing.T) {
	tbl := buildCarryTable(true)

	plan := tbl.PlanLayout(LayoutArea{Width: 200, Height: 80})
	if plan.Status != LayoutPartial {
		t.Fatalf("expected LayoutPartial (split), got %v", plan.Status)
	}

	overflow, ok := plan.Overflow.(*Table)
	if !ok {
		t.Fatalf("expected *Table overflow, got %T", plan.Overflow)
	}

	bf := firstBodyRow(overflow)
	if bf == nil {
		t.Fatal("overflow has no body row")
	}

	if bf.cells[0].text != "Brought forward" {
		t.Fatalf("expected brought-forward row first in continuation, got label %q", bf.cells[0].text)
	}

	p, ok := bf.cells[1].content.(*Paragraph)
	if !ok {
		t.Fatalf("brought-forward value cell lost its element font (got text cell %q)", bf.cells[1].text)
	}

	if p.PlainText() == "" {
		t.Fatal("brought-forward value cell was not filled with the running balance")
	}
}

// TestTableCarryForwardDisabled verifies that without carry config the
// continuation starts with a real data row (no synthesized brought row).
func TestTableCarryForwardDisabled(t *testing.T) {
	tbl := buildCarryTable(false)

	plan := tbl.PlanLayout(LayoutArea{Width: 200, Height: 80})
	if plan.Status != LayoutPartial {
		t.Fatalf("expected LayoutPartial (split), got %v", plan.Status)
	}

	overflow := plan.Overflow.(*Table)

	bf := firstBodyRow(overflow)
	if bf == nil {
		t.Fatal("overflow has no body row")
	}

	if bf.cells[0].text == "Brought forward" {
		t.Fatal("brought-forward row should not appear when carry is unconfigured")
	}
}
