package cybr

import (
	"strconv"
	"testing"
)

type mockOptions struct {
	Bool bool
	Str  string
}

func (m mockOptions) GetDisableHTTPS() bool {
	return m.Bool
}

func (m mockOptions) GetResolvedDomain() string {
	return m.Str
}

func (m mockOptions) GetResolvedSubdomain() string {
	return m.Str
}

func TestGetDisableHTTPS(t *testing.T) {
	cases := []struct {
		Options     []interface{}
		ExpectFound bool
		ExpectValue bool
	}{
		{
			Options: []interface{}{struct{}{}},
		},
		{
			Options: []interface{}{mockOptions{
				Bool: false,
			}},
			ExpectFound: true,
			ExpectValue: false,
		},
		{
			Options: []interface{}{mockOptions{
				Bool: true,
			}},
			ExpectFound: true,
			ExpectValue: true,
		},
		{
			Options:     []interface{}{struct{}{}, mockOptions{Bool: true}, mockOptions{Bool: false}},
			ExpectFound: true,
			ExpectValue: true,
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			value, found := GetDisableHTTPS(tt.Options...)
			if found != tt.ExpectFound {
				t.Fatalf("expect value to not be found")
			}
			if value != tt.ExpectValue {
				t.Errorf("expect %v, got %v", tt.ExpectValue, value)
			}
		})
	}
}

func TestGetResolvedSubdomain(t *testing.T) {
	cases := []struct {
		Options     []interface{}
		ExpectFound bool
		ExpectValue string
	}{
		{
			Options: []interface{}{struct{}{}},
		},
		{
			Options:     []interface{}{mockOptions{Str: ""}},
			ExpectFound: true,
			ExpectValue: "",
		},
		{
			Options:     []interface{}{mockOptions{Str: "foo"}},
			ExpectFound: true,
			ExpectValue: "foo",
		},
		{
			Options:     []interface{}{struct{}{}, mockOptions{Str: "bar"}, mockOptions{Str: "baz"}},
			ExpectFound: true,
			ExpectValue: "bar",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			value, found := GetResolvedSubdomain(tt.Options...)
			if found != tt.ExpectFound {
				t.Fatalf("expect value to not be found")
			}
			if value != tt.ExpectValue {
				t.Errorf("expect %v, got %v", tt.ExpectValue, value)
			}
		})
	}
}

func TestGetResolvedDomain(t *testing.T) {
	cases := []struct {
		Options     []interface{}
		ExpectFound bool
		ExpectValue string
	}{
		{
			Options: []interface{}{struct{}{}},
		},
		{
			Options:     []interface{}{mockOptions{Str: ""}},
			ExpectFound: true,
			ExpectValue: "",
		},
		{
			Options:     []interface{}{mockOptions{Str: "foo"}},
			ExpectFound: true,
			ExpectValue: "foo",
		},
		{
			Options:     []interface{}{struct{}{}, mockOptions{Str: "bar"}, mockOptions{Str: "baz"}},
			ExpectFound: true,
			ExpectValue: "bar",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			value, found := GetResolvedDomain(tt.Options...)
			if found != tt.ExpectFound {
				t.Fatalf("expect value to not be found")
			}
			if value != tt.ExpectValue {
				t.Errorf("expect %v, got %v", tt.ExpectValue, value)
			}
		})
	}
}

var _ EndpointResolverWithOptions = EndpointResolverWithOptionsFunc(nil)

func TestEndpointResolverWithOptionsFunc_ResolveEndpoint(t *testing.T) {
	var er EndpointResolverWithOptions = EndpointResolverWithOptionsFunc(func(subdomain, service, domain string, options ...interface{}) (Endpoint, error) {
		if e, a := "cup", subdomain; e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		if e, a := "foo", service; e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		if e, a := "bar", domain; e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		if e, a := 2, len(options); e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		return Endpoint{
			URL: "https://cup.foo.bar.com",
		}, nil
	})

	e, err := er.ResolveEndpoint("cup", "foo", "bar", 1, 2)
	if err != nil {
		t.Errorf("expect no error, got %v", err)
	}

	if e, a := "https://cup.foo.bar.com", e.URL; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}
