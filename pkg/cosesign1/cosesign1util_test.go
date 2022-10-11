package cosesign1

import (
	_ "embed"
	"testing"

	"github.com/veraison/go-cose"
)

/*
	The inputs here are generated via the Makefile,
	thus if you update the fragment's rego (infra.rego)
	then you can build a matching code file etc by
	make infra.rego.cose
*/

//go:embed infra.rego.base64
var FragmentRego string

//go:embed infra.rego.cose
var FragmentCose []byte

//go:embed infra.rego.cose
var FragmentCose2 []byte

// This is a self signed key which is only used for testing, it is not a risk.
// It enables a check against the key and signature blobs

//go:embed ec384-private.body
var KeyStrippedPem string // Strip off the BEGIN/END so we don't trigger credential checks

var begingPrivateKey = "-----BEGIN PRIVATE KEY-----\n"
var endPrivateKey = "-----END PRIVATE KEY-----"

var KeyPem = begingPrivateKey + KeyStrippedPem + endPrivateKey

//go:embed ec384-cert.crt
var PubCertPem string // the whole cert chain to embed

//go:embed leafcert.ec384-public.pem
var LeafCertPem string // the expected leaf cert

//go:embed leafkey.ec384-public.pem
var LeafKeyPem string

/*
	Decode a COSE_Sign1 document and check that we get the expected payload, issuer, keys, certs etc.
*/

func Test_UnpackAndValidateCannedFragment(t *testing.T) {
	var unpacked UnpackedCoseSign1
	unpacked, err := UnpackAndValidateCOSE1CertChain(FragmentCose, nil, false, false)

	if err != nil {
		t.Errorf("UnpackAndValidateCOSE1CertChain failed: %s", err.Error())
	}
	
	var iss = unpacked.Issuer
	var feed = unpacked.Feed
	var pubkey = unpacked.Pubkey
	var pubcert = unpacked.Pubcert
	var payload = string(unpacked.Payload[:])
	var cty = unpacked.ContentType

	if pubkey != LeafKeyPem && (pubkey+"\n") != LeafKeyPem {
		t.Error("pubkey did not match")
	}
	if pubcert != LeafCertPem && (pubcert+"\n") != LeafCertPem {
		t.Error("pubcert did not match")
	}
	if cty != "application/unknown+json" {
		t.Error("cty did not match")
	}
	if payload != FragmentRego {
		t.Error("payload did not match")
	}
	if iss != "TestIssuer" {
		t.Error("iss did not match")
	}
	if feed != "TestFeed" {
		t.Error("feed did not match")
	}
}

func Test_UnpackAndValidateCannedFragmentCorrupted(t *testing.T) {
	var offset = len(FragmentCose2) / 2
	FragmentCose2[offset] = FragmentCose[offset] + 1 // corrupt the cose document (use the uncorrupted one as source in case we loop back to a good value)
	var _, err = UnpackAndValidateCOSE1CertChain(FragmentCose2, nil, false, false)

	// expect it to fail
	if err == nil {
		t.Error("corrupted document passed validation")
	}
}

/*
	Use CreateCoseSign1 to make a document that should match the one made by the makefile
*/

func Test_CreateCoseSign1Fragment(t *testing.T) {
	var raw, err = CreateCoseSign1([]byte(FragmentRego), "TestIssuer", "TestFeed", "application/unknown+json", []byte(PubCertPem), []byte(KeyPem), "zero", cose.AlgorithmES384, false)
	if err != nil {
		t.Errorf("CreateCoseSign1 failed: %s", err.Error())
	}

	if len(raw) != len(FragmentCose) {
		t.Error("created fragment length does not match expected")
	}

	for which := range raw {
		if raw[which] != FragmentCose[which] {
			t.Errorf("created fragment byte offset %d does not match expected", which)
		}
	}
}
