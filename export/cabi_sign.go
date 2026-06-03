// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

//go:build cgo && !js && !wasm

package main

/*
#include <stdint.h>
*/
import "C"
import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"unsafe"

	"github.com/kikihakiem/folio/sign"
)

// signOptsBuilder accumulates signing options before calling SignPDF.
type signOptsBuilder struct {
	signer      sign.Signer
	level       sign.PAdESLevel
	name        string
	reason      string
	location    string
	contactInfo string
	tsaClient   *sign.TSAClient
	ocspClient  *sign.OCSPClient
}

// ── Signer ─────────────────────────────────────────────────────────

// folio_signer_new_pem creates a signer from PEM-encoded private key and certificate chain.
//
//export folio_signer_new_pem
func folio_signer_new_pem(keyPEM unsafe.Pointer, keyLen C.int32_t,
	certPEM unsafe.Pointer, certLen C.int32_t) C.uint64_t {
	if keyPEM == nil || keyLen <= 0 {
		setLastError("invalid key PEM data")
		return 0
	}
	if certPEM == nil || certLen <= 0 {
		setLastError("invalid cert PEM data")
		return 0
	}

	keyBytes := C.GoBytes(keyPEM, C.int(keyLen))
	certBytes := C.GoBytes(certPEM, C.int(certLen))

	// Parse private key from PEM.
	var privKey crypto.Signer
	for {
		var block *pem.Block
		block, keyBytes = pem.Decode(keyBytes)
		if block == nil {
			break
		}
		// Try PKCS#8 first, then PKCS#1 RSA, then EC.
		if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			switch typed := k.(type) {
			case *rsa.PrivateKey:
				privKey = typed
			case *ecdsa.PrivateKey:
				privKey = typed
			case crypto.Signer:
				privKey = typed
			}
			break
		}
		if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			privKey = k
			break
		}
		if k, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
			privKey = k
			break
		}
	}
	if privKey == nil {
		setLastError("failed to parse private key from PEM")
		return 0
	}

	// Parse certificate chain from PEM.
	var certs []*x509.Certificate
	for {
		var block *pem.Block
		block, certBytes = pem.Decode(certBytes)
		if block == nil {
			break
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		certs = append(certs, cert)
	}
	if len(certs) == 0 {
		setLastError("no certificates found in PEM data")
		return 0
	}

	signer, err := sign.NewLocalSigner(privKey, certs)
	if err != nil {
		setLastError(err.Error())
		return 0
	}
	return C.uint64_t(ht.store(signer))
}

//export folio_signer_free
func folio_signer_free(signerH C.uint64_t) {
	ht.delete(uint64(signerH))
}

// ── TSA / OCSP clients ─────────────────────────────────────────────

//export folio_tsa_client_new
func folio_tsa_client_new(url *C.char) C.uint64_t {
	tsa := sign.NewTSAClient(C.GoString(url))
	return C.uint64_t(ht.store(tsa))
}

//export folio_tsa_client_free
func folio_tsa_client_free(tsaH C.uint64_t) {
	ht.delete(uint64(tsaH))
}

//export folio_ocsp_client_new
func folio_ocsp_client_new() C.uint64_t {
	return C.uint64_t(ht.store(sign.NewOCSPClient()))
}

//export folio_ocsp_client_free
func folio_ocsp_client_free(ocspH C.uint64_t) {
	ht.delete(uint64(ocspH))
}

// ── Sign options builder ───────────────────────────────────────────

//export folio_sign_opts_new
func folio_sign_opts_new(signerH C.uint64_t, level C.int32_t) C.uint64_t {
	s, errCode := loadSignSigner(signerH)
	if errCode != errOK {
		return 0
	}
	opts := &signOptsBuilder{
		signer: s,
		level:  sign.PAdESLevel(level),
	}
	return C.uint64_t(ht.store(opts))
}

//export folio_sign_opts_set_name
func folio_sign_opts_set_name(optsH C.uint64_t, name *C.char) C.int32_t {
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return errCode
	}
	opts.name = C.GoString(name)
	return errOK
}

//export folio_sign_opts_set_reason
func folio_sign_opts_set_reason(optsH C.uint64_t, reason *C.char) C.int32_t {
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return errCode
	}
	opts.reason = C.GoString(reason)
	return errOK
}

//export folio_sign_opts_set_location
func folio_sign_opts_set_location(optsH C.uint64_t, location *C.char) C.int32_t {
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return errCode
	}
	opts.location = C.GoString(location)
	return errOK
}

//export folio_sign_opts_set_contact_info
func folio_sign_opts_set_contact_info(optsH C.uint64_t, info *C.char) C.int32_t {
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return errCode
	}
	opts.contactInfo = C.GoString(info)
	return errOK
}

//export folio_sign_opts_set_tsa
func folio_sign_opts_set_tsa(optsH C.uint64_t, tsaH C.uint64_t) C.int32_t {
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return errCode
	}
	tsa, errCode := loadTSAClient(tsaH)
	if errCode != errOK {
		return errCode
	}
	opts.tsaClient = tsa
	return errOK
}

//export folio_sign_opts_set_ocsp
func folio_sign_opts_set_ocsp(optsH C.uint64_t, ocspH C.uint64_t) C.int32_t {
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return errCode
	}
	ocsp, errCode := loadOCSPClient(ocspH)
	if errCode != errOK {
		return errCode
	}
	opts.ocspClient = ocsp
	return errOK
}

//export folio_sign_opts_free
func folio_sign_opts_free(optsH C.uint64_t) {
	ht.delete(uint64(optsH))
}

// ── Sign PDF ───────────────────────────────────────────────────────

// folio_sign_pdf signs a PDF and returns the signed bytes as a buffer handle.
//
//export folio_sign_pdf
func folio_sign_pdf(pdfData unsafe.Pointer, pdfLen C.int32_t, optsH C.uint64_t) C.uint64_t {
	if pdfData == nil || pdfLen <= 0 {
		setLastError("invalid PDF data")
		return 0
	}
	opts, errCode := loadSignOpts(optsH)
	if errCode != errOK {
		return 0
	}
	pdfBytes := C.GoBytes(pdfData, C.int(pdfLen))
	signOpts := sign.Options{
		Signer:      opts.signer,
		Level:       opts.level,
		Name:        opts.name,
		Reason:      opts.reason,
		Location:    opts.location,
		ContactInfo: opts.contactInfo,
		TSAClient:   opts.tsaClient,
		OCSPClient:  opts.ocspClient,
	}
	signed, err := sign.SignPDF(pdfBytes, signOpts)
	if err != nil {
		setLastError(err.Error())
		return 0
	}
	return C.uint64_t(ht.store(newCBuffer(signed)))
}

// ── Helpers ────────────────────────────────────────────────────────

func loadSignSigner(h C.uint64_t) (sign.Signer, C.int32_t) {
	v := ht.load(uint64(h))
	if v == nil {
		setLastError("invalid signer handle")
		return nil, errInvalidHandle
	}
	s, ok := v.(sign.Signer)
	if !ok {
		setLastError(fmt.Sprintf("handle %d is not a signer (type %T)", uint64(h), v))
		return nil, errTypeMismatch
	}
	return s, errOK
}

func loadSignOpts(h C.uint64_t) (*signOptsBuilder, C.int32_t) {
	v := ht.load(uint64(h))
	if v == nil {
		setLastError("invalid sign opts handle")
		return nil, errInvalidHandle
	}
	opts, ok := v.(*signOptsBuilder)
	if !ok {
		setLastError(fmt.Sprintf("handle %d is not sign opts (type %T)", uint64(h), v))
		return nil, errTypeMismatch
	}
	return opts, errOK
}

func loadTSAClient(h C.uint64_t) (*sign.TSAClient, C.int32_t) {
	v := ht.load(uint64(h))
	if v == nil {
		setLastError("invalid TSA client handle")
		return nil, errInvalidHandle
	}
	tsa, ok := v.(*sign.TSAClient)
	if !ok {
		setLastError(fmt.Sprintf("handle %d is not a TSA client (type %T)", uint64(h), v))
		return nil, errTypeMismatch
	}
	return tsa, errOK
}

func loadOCSPClient(h C.uint64_t) (*sign.OCSPClient, C.int32_t) {
	v := ht.load(uint64(h))
	if v == nil {
		setLastError("invalid OCSP client handle")
		return nil, errInvalidHandle
	}
	ocsp, ok := v.(*sign.OCSPClient)
	if !ok {
		setLastError(fmt.Sprintf("handle %d is not an OCSP client (type %T)", uint64(h), v))
		return nil, errTypeMismatch
	}
	return ocsp, errOK
}
