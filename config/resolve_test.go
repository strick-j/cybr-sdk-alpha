package config

import (
	"context"
	"testing"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/cybr-sdk-alpha/internal/cybrtesting/unit"

	"github.com/strick-j/smithy-go/logging"
)

func TestGetResolvedSubdomain(t *testing.T) {
	var options LoadOptions
	optFns := []func(options *LoadOptions) error{
		WithSubdomain("ignored-subdomain"),

		WithSubdomain("bar"),
	}

	for _, optFn := range optFns {
		optFn(&options)
	}

	configs := configs{options}

	var cfg cybr.Config

	if err := resolveSubdomain(context.Background(), &cfg, configs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if e, a := "bar", cfg.SubDomain; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestGetResolvedDomain(t *testing.T) {
	var options LoadOptions
	optFns := []func(options *LoadOptions) error{
		WithDomain("ignored-domain"),

		WithDomain("com"),
	}

	for _, optFn := range optFns {
		optFn(&options)
	}

	configs := configs{options}

	var cfg cybr.Config

	if err := resolveDomain(context.Background(), &cfg, configs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if e, a := "com", cfg.Domain; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestDefaultDomain(t *testing.T) {
	ctx := context.Background()

	var options LoadOptions
	WithDefaultDomain("foo-domain")(&options)

	configs := configs{options}
	cfg := unit.Config()

	err := resolveDefaultDomain(ctx, &cfg, configs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if e, a := "mock-domain", cfg.Domain; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	cfg.Domain = ""

	err = resolveDefaultDomain(ctx, &cfg, configs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if e, a := "foo-domain", cfg.Domain; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestResolveLogger(t *testing.T) {
	cfg, err := LoadDefaultConfig(context.Background(), func(o *LoadOptions) error {
		o.Logger = logging.Nop{}
		return nil
	})
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	_, ok := cfg.Logger.(logging.Nop)
	if !ok {
		t.Error("unexpected logger type")
	}
}

func TestEndpointResolverWithOptionsFunc_ResolveEndpoint(t *testing.T) {

}
