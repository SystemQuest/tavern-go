package regex

import (
	"bytes"
	"testing"
)

func TestValidate_SimpleMatch(t *testing.T) {
	result, err := Validate("hello world", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected no named groups, got %v", result)
	}
}

func TestValidate_NamedGroups(t *testing.T) {
	result, err := Validate("token=abc123", `token=(?P<token>\w+)`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["token"] != "abc123" {
		t.Errorf("expected token=abc123, got %v", result["token"])
	}
}

func TestValidate_MultipleGroups(t *testing.T) {
	result, err := Validate(
		"<a href=\"http://localhost/verify?token=xyz\">",
		`<a href="(?P<url>.*?)\?token=(?P<token>.*?)">`,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["url"] != "http://localhost/verify" {
		t.Errorf("expected url, got %v", result["url"])
	}
	if result["token"] != "xyz" {
		t.Errorf("expected token=xyz, got %v", result["token"])
	}
}

func TestValidate_NoMatch(t *testing.T) {
	_, err := Validate("hello", "goodbye")
	if err == nil {
		t.Fatal("expected error for no match")
	}
}

func TestValidate_InvalidRegex(t *testing.T) {
	_, err := Validate("hello", "[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestValidate_EmptyExpression(t *testing.T) {
	_, err := Validate("hello", "")
	if err == nil {
		t.Fatal("expected error for empty expression")
	}
}

func TestValidateReader_Success(t *testing.T) {
	reader := bytes.NewReader([]byte("token=abc123"))
	result, err := ValidateReader(reader, `token=(?P<token>\w+)`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["token"] != "abc123" {
		t.Errorf("expected token=abc123, got %v", result["token"])
	}
}
