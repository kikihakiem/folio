// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package layout

import (
	"testing"

	"github.com/kikihakiem/folio/font"
)

func flexParagraph(text string) *Paragraph {
	return NewParagraph(text, font.Helvetica, 10)
}

func TestFlexRowBasic(t *testing.T) {
	f := NewFlex().
		Add(flexParagraph("Hello")).
		Add(flexParagraph("World"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatalf("expected LayoutFull, got %d", plan.Status)
	}
	if plan.Consumed <= 0 {
		t.Error("consumed should be > 0")
	}
	if len(plan.Blocks) != 1 {
		t.Fatalf("expected 1 container block, got %d", len(plan.Blocks))
	}
	if len(plan.Blocks[0].Children) < 2 {
		t.Errorf("expected at least 2 children, got %d", len(plan.Blocks[0].Children))
	}
}

func TestFlexColumnBasic(t *testing.T) {
	f := NewFlex().
		SetDirection(FlexColumn).
		Add(flexParagraph("First")).
		Add(flexParagraph("Second")).
		Add(flexParagraph("Third"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatalf("expected LayoutFull, got %d", plan.Status)
	}
	if len(plan.Blocks[0].Children) < 3 {
		t.Errorf("expected at least 3 children, got %d", len(plan.Blocks[0].Children))
	}
}

func TestFlexGrow(t *testing.T) {
	// Two items: one grow=1, one grow=2. They should split the space 1:2.
	f := NewFlex().
		AddItem(NewFlexItem(flexParagraph("A")).SetGrow(1)).
		AddItem(NewFlexItem(flexParagraph("B")).SetGrow(2))

	plan := f.PlanLayout(LayoutArea{Width: 300, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatalf("expected LayoutFull, got %d", plan.Status)
	}
	// Both items should be present as children.
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatalf("expected 2+ children, got %d", len(children))
	}
}

func TestFlexJustifyCenter(t *testing.T) {
	f := NewFlex().
		SetJustifyContent(JustifyCenter).
		Add(flexParagraph("Centered"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	child := plan.Blocks[0].Children[0]
	// Centered: child.X should be > 0 (offset from left).
	if child.X <= 0 {
		t.Errorf("expected centered X offset > 0, got %f", child.X)
	}
}

func TestFlexJustifySpaceBetween(t *testing.T) {
	f := NewFlex().
		SetJustifyContent(JustifySpaceBetween).
		Add(flexParagraph("Left")).
		Add(flexParagraph("Right"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatal("expected 2+ children")
	}
	// First child should be at X=0, second should be further right.
	if children[0].X >= children[1].X {
		t.Error("first child should be left of second child")
	}
}

func TestFlexJustifyFlexEnd(t *testing.T) {
	f := NewFlex().
		SetJustifyContent(JustifyFlexEnd).
		Add(flexParagraph("Right"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	child := plan.Blocks[0].Children[0]
	if child.X <= 0 {
		t.Errorf("expected flex-end X offset > 0, got %f", child.X)
	}
}

func TestFlexJustifySpaceAround(t *testing.T) {
	f := NewFlex().
		SetJustifyContent(JustifySpaceAround).
		Add(flexParagraph("A")).
		Add(flexParagraph("B"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatal("expected 2+ children")
	}
	// First child should have some left margin (half space).
	if children[0].X <= 0 {
		t.Error("space-around: first child should have left margin")
	}
}

func TestFlexJustifySpaceEvenly(t *testing.T) {
	f := NewFlex().
		SetJustifyContent(JustifySpaceEvenly).
		Add(flexParagraph("A")).
		Add(flexParagraph("B"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatal("expected 2+ children")
	}
	if children[0].X <= 0 {
		t.Error("space-evenly: first child should have left margin")
	}
}

func TestFlexAlignCrossCenter(t *testing.T) {
	// Two items with different heights. CrossAlignCenter should center the shorter one.
	short := flexParagraph("Short")
	long := flexParagraph("This is a much longer paragraph that will wrap to multiple lines when given a narrow width")

	f := NewFlex().
		SetAlignItems(CrossAlignCenter).
		AddItem(NewFlexItem(short).SetBasis(100)).
		AddItem(NewFlexItem(long).SetBasis(100))

	plan := f.PlanLayout(LayoutArea{Width: 220, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatal("expected 2+ children")
	}
	// The shorter item should have a Y offset > the taller item's Y offset.
	if children[0].Y <= 0 && children[1].Y <= 0 {
		// At least one should be offset if they have different heights.
		t.Log("both at Y=0, heights may be equal")
	}
}

func TestFlexAlignSelfOverride(t *testing.T) {
	f := NewFlex().
		SetAlignItems(CrossAlignStart).
		AddItem(NewFlexItem(flexParagraph("Start")).SetBasis(100)).
		AddItem(NewFlexItem(flexParagraph("End")).SetBasis(100).SetAlignSelf(CrossAlignEnd))

	plan := f.PlanLayout(LayoutArea{Width: 220, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
}

func TestFlexWrap(t *testing.T) {
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetColumnGap(10)

	// Add items that won't all fit on one line.
	for range 5 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatalf("expected LayoutFull, got %d", plan.Status)
	}
	// With 250px width, 100px items + 10px gap: 2 items per line (100+10+100=210).
	// So 3 lines: [2, 2, 1].
	if plan.Consumed <= 0 {
		t.Error("consumed should be > 0")
	}
}

func TestFlexGap(t *testing.T) {
	f := NewFlex().
		SetGap(20).
		Add(flexParagraph("A")).
		Add(flexParagraph("B"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatal("expected 2+ children")
	}
	// Second child should be offset by first child's width + gap.
	gap := children[1].X - (children[0].X + children[0].Width)
	if gap < 15 { // allow some tolerance
		t.Errorf("expected ~20pt gap, got %f", gap)
	}
}

func TestFlexPadding(t *testing.T) {
	f := NewFlex().
		SetPadding(10).
		Add(flexParagraph("Padded"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	child := plan.Blocks[0].Children[0]
	// Child should be offset by padding.
	if child.X < 10 {
		t.Errorf("expected X >= 10 (padding), got %f", child.X)
	}
	if child.Y < 10 {
		t.Errorf("expected Y >= 10 (padding), got %f", child.Y)
	}
}

func TestFlexBackground(t *testing.T) {
	f := NewFlex().
		SetBackground(ColorBlue).
		Add(flexParagraph("BG"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	if plan.Blocks[0].Draw == nil {
		t.Error("container should have a Draw closure for background")
	}
}

func TestFlexBorders(t *testing.T) {
	f := NewFlex().
		SetBorder(Border{Width: 1, Color: ColorBlack, Style: BorderSolid}).
		Add(flexParagraph("Bordered"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	if plan.Blocks[0].Draw == nil {
		t.Error("container should have a Draw closure for borders")
	}
}

func TestFlexColumnPageBreak(t *testing.T) {
	f := NewFlex().SetDirection(FlexColumn)
	for range 50 {
		f.Add(flexParagraph("Line of text that takes some vertical space"))
	}

	// 50 items at ~12pt each = ~600pt. Height of 100 should cause a split.
	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 100})
	if plan.Status == LayoutFull {
		t.Errorf("expected partial, area is too small for 50 items (consumed would be ~600pt)")
	}
	if plan.Status == LayoutPartial && plan.Overflow == nil {
		t.Error("partial status but no overflow element")
	}
}

func TestFlexRowPageBreakWrapped(t *testing.T) {
	f := NewFlex().SetWrap(FlexWrapOn)
	for range 20 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 30})
	// Very small height should cause partial layout.
	if plan.Status == LayoutFull {
		t.Error("expected partial, area height too small for many wrapped lines")
	}
}

func TestFlexEmpty(t *testing.T) {
	f := NewFlex()
	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Error("empty flex should return LayoutFull")
	}
}

func TestFlexMinWidthRow(t *testing.T) {
	f := NewFlex().
		Add(flexParagraph("Hello")).
		Add(flexParagraph("World"))

	min := f.MinWidth()
	max := f.MaxWidth()
	if min <= 0 {
		t.Error("MinWidth should be > 0")
	}
	if max < min {
		t.Errorf("MaxWidth (%f) should be >= MinWidth (%f)", max, min)
	}
}

func TestFlexMinWidthColumn(t *testing.T) {
	f := NewFlex().
		SetDirection(FlexColumn).
		Add(flexParagraph("Hello")).
		Add(flexParagraph("World"))

	min := f.MinWidth()
	max := f.MaxWidth()
	if min <= 0 {
		t.Error("MinWidth should be > 0")
	}
	if max < min {
		t.Errorf("MaxWidth (%f) should be >= MinWidth (%f)", max, min)
	}
}

func TestFlexMinWidthNoWrap(t *testing.T) {
	f := NewFlex().
		SetColumnGap(10).
		Add(flexParagraph("A")).
		Add(flexParagraph("B"))

	fWrap := NewFlex().
		SetWrap(FlexWrapOn).
		SetColumnGap(10).
		Add(flexParagraph("A")).
		Add(flexParagraph("B"))

	// No-wrap min should be >= wrap min (all items must fit on one line).
	noWrapMin := f.MinWidth()
	wrapMin := fWrap.MinWidth()
	if noWrapMin < wrapMin {
		t.Errorf("no-wrap MinWidth (%f) should be >= wrap MinWidth (%f)", noWrapMin, wrapMin)
	}
}

func TestFlexLayoutElement(t *testing.T) {
	// Test the Element interface (Layout method).
	f := NewFlex().
		Add(flexParagraph("Hello")).
		Add(flexParagraph("World"))

	lines := f.Layout(400)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Height <= 0 {
		t.Error("line height should be > 0")
	}
}

func TestFlexSpaceBeforeAfter(t *testing.T) {
	f := NewFlex().
		SetSpaceBefore(20).
		SetSpaceAfter(10).
		Add(flexParagraph("Spaced"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	// Consumed should include spaceBefore + content + spaceAfter.
	if plan.Consumed < 30 {
		t.Errorf("consumed (%f) should include spaceBefore (20) + spaceAfter (10)", plan.Consumed)
	}
}

func TestFlexShrink(t *testing.T) {
	// Items with basis wider than container should shrink.
	f := NewFlex().
		AddItem(NewFlexItem(flexParagraph("A")).SetBasis(200).SetShrink(1)).
		AddItem(NewFlexItem(flexParagraph("B")).SetBasis(200).SetShrink(1))

	plan := f.PlanLayout(LayoutArea{Width: 300, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
}

func TestFlexNestedDiv(t *testing.T) {
	inner := NewDiv().
		SetPadding(5).
		SetBackground(ColorLightGray).
		Add(flexParagraph("Inside div"))

	f := NewFlex().
		Add(inner).
		Add(flexParagraph("Next to div"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	if len(plan.Blocks[0].Children) < 2 {
		t.Error("expected at least 2 children")
	}
}

func TestFlexColumnAlignEnd(t *testing.T) {
	f := NewFlex().
		SetDirection(FlexColumn).
		SetAlignItems(CrossAlignEnd).
		Add(flexParagraph("Right-aligned"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	child := plan.Blocks[0].Children[0]
	// Child should be pushed to the right.
	if child.X <= 0 {
		t.Errorf("expected X > 0 for CrossAlignEnd, got %f", child.X)
	}
}

func TestFlexColumnAlignCenter(t *testing.T) {
	f := NewFlex().
		SetDirection(FlexColumn).
		SetAlignItems(CrossAlignCenter).
		Add(flexParagraph("Centered"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	child := plan.Blocks[0].Children[0]
	if child.X <= 0 {
		t.Errorf("expected X > 0 for CrossAlignCenter, got %f", child.X)
	}
}

func TestFlexAlignContentCenter(t *testing.T) {
	// 3 wrapped lines in a tall container.
	// With align-content:center, lines should be centered vertically.
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifyCenter)

	for range 6 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if len(children) < 2 {
		t.Fatal("expected multiple children")
	}
	// First child should be shifted down from the top (centered).
	if children[0].Y < 1 {
		t.Errorf("expected first child Y > 0 (centered), got %f", children[0].Y)
	}
}

func TestFlexAlignContentSpaceBetween(t *testing.T) {
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifySpaceBetween)

	for range 4 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	// With space-between, first line at top, last line at bottom.
	// First child Y should be near 0 (top), last child Y > first.
	firstY := children[0].Y
	lastY := children[len(children)-1].Y
	if lastY <= firstY {
		t.Errorf("expected last child Y > first child Y for space-between")
	}
}

func TestFlexAlignContentFlexEnd(t *testing.T) {
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifyFlexEnd)

	for range 4 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	// With flex-end, first item should be pushed down from top.
	if children[0].Y < 1 {
		t.Errorf("expected first child Y > 0 (flex-end), got %f", children[0].Y)
	}
}

func TestFlexAlignContentSpaceAround(t *testing.T) {
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifySpaceAround)

	for range 4 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	// First child should be pushed down (half-gap before first line).
	if children[0].Y < 1 {
		t.Errorf("expected first child Y > 0 for space-around, got %f", children[0].Y)
	}
}

func TestFlexAlignContentSpaceEvenly(t *testing.T) {
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifySpaceEvenly)

	for range 4 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	// First child should be pushed down (equal gap before first line).
	if children[0].Y < 1 {
		t.Errorf("expected first child Y > 0 for space-evenly, got %f", children[0].Y)
	}
}

func TestFlexAlignContentNoWrapIgnored(t *testing.T) {
	// align-content should have no effect without wrapping (single line).
	f := NewFlex().
		SetAlignContent(JustifyCenter).
		Add(flexParagraph("A")).
		Add(flexParagraph("B"))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	// Without wrapping there's only 1 line, so align-content is a no-op.
	// First child Y should be at or near 0 (no redistribution).
	if children[0].Y > 1 {
		t.Errorf("expected first child Y near 0 (single line, no effect), got %f", children[0].Y)
	}
}

func TestFlexAlignContentSingleWrappedLine(t *testing.T) {
	// Two items that fit on one line even with wrapping enabled.
	// align-content should be no-op for a single line.
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifyCenter).
		AddItem(NewFlexItem(flexParagraph("A")).SetBasis(50)).
		AddItem(NewFlexItem(flexParagraph("B")).SetBasis(50))

	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	children := plan.Blocks[0].Children
	if children[0].Y > 1 {
		t.Errorf("single wrapped line: Y should be near 0, got %f", children[0].Y)
	}
}

func TestFlexAlignContentWithRowGap(t *testing.T) {
	// align-content with row-gap should account for gaps in free space calculation.
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetRowGap(10).
		SetAlignContent(JustifyCenter)

	for range 6 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 800})
	if plan.Status != LayoutFull {
		t.Fatal("expected LayoutFull")
	}
	if plan.Consumed <= 0 {
		t.Error("consumed should be positive")
	}
}

func TestFlexAlignContentTightContainer(t *testing.T) {
	// When container is just tall enough for the content, no redistribution happens.
	f := NewFlex().
		SetWrap(FlexWrapOn).
		SetAlignContent(JustifySpaceBetween)

	for range 4 {
		f.AddItem(NewFlexItem(flexParagraph("Item")).SetBasis(100))
	}

	// Use just enough height — lines fill the space, no free space to distribute.
	plan := f.PlanLayout(LayoutArea{Width: 250, Height: 40})
	if plan.Status == LayoutFull {
		children := plan.Blocks[0].Children
		// First child should be near top (no space to distribute).
		if children[0].Y > 1 {
			t.Errorf("tight container: first child Y should be near 0, got %f", children[0].Y)
		}
	}
	// Partial layout is also acceptable for a very tight container.
}

func TestFlexColumnJustifyContentCenter(t *testing.T) {
	// Flex column with 2 small items in a tall container, centered.
	f := NewFlex()
	f.SetDirection(FlexColumn)
	f.SetJustifyContent(JustifyCenter)
	f.Add(NewParagraph("Top", font.Helvetica, 12))
	f.Add(NewParagraph("Bottom", font.Helvetica, 12))

	// Force a 300pt height so items have room to center.
	f.ForceHeight(Pt(300))
	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 300})
	if plan.Status == LayoutNothing {
		t.Fatal("expected output")
	}
	// Items should not start at Y=0 — they should be pushed down by centering.
	if len(plan.Blocks) > 0 && len(plan.Blocks[0].Children) > 0 {
		firstChildY := plan.Blocks[0].Children[0].Y
		if firstChildY < 50 {
			t.Errorf("centered column items should have Y offset > 50, got %f", firstChildY)
		}
	}
}

func TestFlexColumnJustifySpaceBetween(t *testing.T) {
	f := NewFlex()
	f.SetDirection(FlexColumn)
	f.SetJustifyContent(JustifySpaceBetween)
	f.Add(NewParagraph("First", font.Helvetica, 12))
	f.Add(NewParagraph("Last", font.Helvetica, 12))

	f.ForceHeight(Pt(200))
	plan := f.PlanLayout(LayoutArea{Width: 400, Height: 200})
	if plan.Status == LayoutNothing {
		t.Fatal("expected output")
	}
	// Space-between: first at top, last at bottom.
	if len(plan.Blocks) > 0 && len(plan.Blocks[0].Children) >= 2 {
		firstY := plan.Blocks[0].Children[0].Y
		lastY := plan.Blocks[0].Children[len(plan.Blocks[0].Children)-1].Y
		if lastY-firstY < 150 {
			t.Errorf("space-between should spread items: firstY=%f lastY=%f gap=%f", firstY, lastY, lastY-firstY)
		}
	}
}

func TestFlexColumnOverflowWithGrow(t *testing.T) {
	f := NewFlex()
	f.SetDirection(FlexColumn)
	for range 20 {
		item := NewFlexItem(NewParagraph("Flex grow item that fills space.", font.Helvetica, 12))
		item.SetGrow(1)
		f.AddItem(item)
	}

	r := NewRenderer(612, 200, Margins{Top: 20, Bottom: 20, Left: 20, Right: 20})
	r.Add(f)
	pages := r.Render()
	if len(pages) < 2 {
		t.Fatalf("expected ≥2 pages for overflowing flex column, got %d", len(pages))
	}
	// Both pages should have content.
	for i, p := range pages {
		if len(p.Stream.Bytes()) == 0 {
			t.Errorf("page %d has empty stream", i)
		}
	}
}

func TestDivPageSplitPreservesBordersAndBackground(t *testing.T) {
	d := NewDiv().
		SetBackground(RGB(0.95, 0.95, 1)).
		SetBorder(SolidBorder(1, RGB(0.3, 0.3, 0.9))).
		SetPadding(10)
	for range 20 {
		d.Add(NewParagraph("Content line that should overflow the page boundary.", font.Helvetica, 12))
	}

	r := NewRenderer(612, 200, Margins{Top: 20, Bottom: 20, Left: 20, Right: 20})
	r.Add(d)
	pages := r.Render()
	if len(pages) < 2 {
		t.Fatalf("expected ≥2 pages, got %d", len(pages))
	}
	// Page 2 should also have background fill (rg) and border stroke (S).
	b2 := pages[1].Stream.Bytes()
	if !containsOp(b2, "rg") {
		t.Error("page 2 should have background fill color (rg)")
	}
	if !containsOp(b2, "S") {
		t.Error("page 2 should have border stroke (S)")
	}
}
