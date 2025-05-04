package normalizer_test

import (
	"testing"
	"time"

	"normalizer/internal/normalizer"
	"slices"
)

func contains(slice []string, val string) bool {
	return slices.Contains(slice, val)
}

func TestExtractIOCs_SingleIOCInputs(t *testing.T) {
	ip := "8.8.8.8"
	domain := "malicious-domain.com"
	sha256 := "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"

	iocsIP := normalizer.ExtractIOCs(ip)
	iocsDomain := normalizer.ExtractIOCs(domain)
	iocsHash := normalizer.ExtractIOCs(sha256)

	if !contains(iocsIP, ip) {
		t.Errorf("Expected IP %q found in IOCs", ip)
	}
	if !contains(iocsDomain, domain) {
		t.Errorf("Expected domain %q found in IOCs", domain)
	}
	if !contains(iocsHash, sha256) {
		t.Errorf("Expected hash %q found in IOCs", sha256)
	}
}

func TestExtractIOCs_CombinedStringWithSpaces(t *testing.T) {
	// Only spaces between tokens to satisfy word boundaries
	input := "192.168.1.1 example.com abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"

	iocs := normalizer.ExtractIOCs(input)

	expected := []string{
		"192.168.1.1",
		"example.com",
		"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	for _, exp := range expected {
		if !contains(iocs, exp) {
			t.Errorf("Expected '%s' in extracted IOCs, got %v", exp, iocs)
		}
	}
}

func TestExtractIOCs_HashWithSpaces(t *testing.T) {
	hash := "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"

	// Hash surrounded by spaces should be matched
	input := " " + hash + " "
	iocs := normalizer.ExtractIOCs(input)
	if !contains(iocs, hash) {
		t.Errorf("Expected hash %q found when surrounded by spaces, got %v", hash, iocs)
	}
}

func TestCreateSTIXBundle_Basic(t *testing.T) {
	processor := normalizer.NewProcessor(nil, nil)
	iocs := []string{"1.2.3.4"}
	tm := time.Now().UTC()

	bundle := processor.CreateSTIXBundle(iocs, tm)
	if bundle == nil {
		t.Fatal("Expected non-nil bundle")
	}
	if len(bundle.Objects) != 4 {
		t.Errorf("Expected 4 objects for one IOC, got %d", len(bundle.Objects))
	}
}
