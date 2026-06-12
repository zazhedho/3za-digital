package database

import "testing"

func TestWithPostgresTimeZoneAddsTimeZoneToDatabaseURL(t *testing.T) {
	got := withPostgresTimeZone("postgres://user:pass@localhost:5432/app?sslmode=require")
	want := "postgres://user:pass@localhost:5432/app?sslmode=require&TimeZone=Asia%2FJakarta"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWithPostgresTimeZoneKeepsExistingTimeZone(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/app?sslmode=require&TimeZone=UTC"

	if got := withPostgresTimeZone(dsn); got != dsn {
		t.Fatalf("expected %q, got %q", dsn, got)
	}
}

func TestWithPostgresTimeZoneAddsKeywordTimeZone(t *testing.T) {
	got := withPostgresTimeZone("host=localhost dbname=app sslmode=disable")
	want := "host=localhost dbname=app sslmode=disable TimeZone=Asia/Jakarta"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
