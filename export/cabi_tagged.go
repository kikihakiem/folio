// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

//go:build cgo && !js && !wasm

package main

/*
#include <stdint.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"

	"github.com/kikihakiem/folio/layout"
)

// ── Alt text for figure elements ───────────────────────────────────

//export folio_image_element_set_alt_text
func folio_image_element_set_alt_text(ieH C.uint64_t, text *C.char) C.int32_t {
	v := ht.load(uint64(ieH))
	if v == nil {
		setLastError("invalid image element handle")
		return errInvalidHandle
	}
	ie, ok := v.(*layout.ImageElement)
	if !ok {
		setLastError(fmt.Sprintf("handle %d is not an image element", uint64(ieH)))
		return errTypeMismatch
	}
	ie.SetAltText(C.GoString(text))
	return errOK
}

//export folio_barcode_element_set_alt_text
func folio_barcode_element_set_alt_text(beH C.uint64_t, text *C.char) C.int32_t {
	be, errCode := loadBarcodeElement(beH)
	if errCode != errOK {
		return errCode
	}
	be.SetAltText(C.GoString(text))
	return errOK
}

//export folio_svg_element_set_alt_text
func folio_svg_element_set_alt_text(seH C.uint64_t, text *C.char) C.int32_t {
	se, errCode := loadSVGElement(seH)
	if errCode != errOK {
		return errCode
	}
	se.SetAltText(C.GoString(text))
	return errOK
}

// ── Custom structure tags ──────────────────────────────────────────

//export folio_div_set_tag
func folio_div_set_tag(divH C.uint64_t, tag *C.char) C.int32_t {
	div, errCode := loadDiv(divH)
	if errCode != errOK {
		return errCode
	}
	div.SetTag(C.GoString(tag))
	return errOK
}

// ── Structure tree reading ─────────────────────────────────────────

// folio_reader_structure_tree returns the document's structure tree as JSON.
// Returns 0 if the document is not tagged.
//
//export folio_reader_structure_tree
func folio_reader_structure_tree(rH C.uint64_t) C.uint64_t {
	r, errCode := loadReader(rH)
	if errCode != errOK {
		return 0
	}
	tree := r.StructureTree()
	if tree == nil {
		setLastError("document is not tagged")
		return 0
	}
	data, err := json.Marshal(tree)
	if err != nil {
		setLastError(err.Error())
		return 0
	}
	return C.uint64_t(ht.store(newCBuffer(data)))
}
