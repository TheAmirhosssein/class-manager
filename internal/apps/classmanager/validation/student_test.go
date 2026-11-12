package validation

import "testing"

func TestValidNationalCode(t *testing.T) {
	if !ValidNationalCode("") {
		t.Fatal("expected empty national code to be valid")
	}
	if !ValidNationalCode("0010350829") {
		t.Fatal("expected valid national code")
	}
	if ValidNationalCode("1111111111") {
		t.Fatal("expected repeated digits to be invalid")
	}
	if ValidNationalCode("123") {
		t.Fatal("expected short code to be invalid")
	}
	if !ValidNationalCode("۰۰۱۰۳۵۰۸۲۹") {
		t.Fatal("expected persian digits to normalize and validate")
	}
}
