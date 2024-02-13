package credentials

import (
	"context"
	"testing"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
)

func TestStaticCredentialsProvider(t *testing.T) {
	s := StaticCredentialsProvider{
		Value: cybr.Credentials{
			Username:     "USERNAME",
			Password:     "PASSWORD",
			SessionToken: "",
		},
	}

	creds, err := s.Retrieve(context.Background())
	if err != nil {
		t.Errorf("expect no error, got %v", err)
	}
	if e, a := "USERNAME", creds.Username; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := "PASSWORD", creds.Password; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if l := creds.SessionToken; len(l) != 0 {
		t.Errorf("expect no token, got %v", l)
	}
}

func TestStaticCredentialsProviderIsExpired(t *testing.T) {
	s := StaticCredentialsProvider{
		Value: cybr.Credentials{
			Username:     "USERNAME",
			Password:     "PASSWORD",
			SessionToken: "",
		},
	}

	creds, err := s.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	if creds.Expired() {
		t.Errorf("expect static credentials to never expire")
	}
}
