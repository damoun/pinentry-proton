package protonpass

import (
	"context"
	"strings"
	"testing"
)

// BenchmarkRetrievePassword benchmarks password retrieval with mock
func BenchmarkRetrievePassword(b *testing.B) {
	mockCLI := createMockCLI(b, "vault", "item", "password", "test-password-123")
	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		password, err := client.RetrievePassword(ctx, "pass://vault/item/password")
		if err != nil {
			b.Fatalf("RetrievePassword failed: %v", err)
		}
		ZeroBytes(password)
	}
}

// BenchmarkRetrievePassword_LongPassword benchmarks with 1KB password
func BenchmarkRetrievePassword_LongPassword(b *testing.B) {
	longPassword := strings.Repeat("a", 1024)
	mockCLI := createMockCLI(b, "vault", "item", "password", longPassword)
	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		password, err := client.RetrievePassword(ctx, "pass://vault/item/password")
		if err != nil {
			b.Fatalf("RetrievePassword failed: %v", err)
		}
		ZeroBytes(password)
	}
}

// BenchmarkRetrievePassword_SpecialChars benchmarks with special characters
func BenchmarkRetrievePassword_SpecialChars(b *testing.B) {
	specialPassword := "p@ss!w#rd$%^&*()[]{}|\\:;\"'<>,.?/~`"
	mockCLI := createMockCLI(b, "vault", "item", "password", specialPassword)
	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		password, err := client.RetrievePassword(ctx, "pass://vault/item/password")
		if err != nil {
			b.Fatalf("RetrievePassword failed: %v", err)
		}
		ZeroBytes(password)
	}
}

// BenchmarkURIParsing benchmarks URI parsing with different formats
func BenchmarkURIParsing(b *testing.B) {
	// Use a simple mock that matches any vault/item
	mockCLI := createMockCLI(b, "vault", "item", "password", "test")
	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()

	// Test just one URI repeatedly to isolate parsing performance
	uri := "pass://vault/item/password"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		password, err := client.RetrievePassword(ctx, uri)
		if err != nil {
			b.Fatalf("RetrievePassword failed: %v", err)
		}
		ZeroBytes(password)
	}
}

// BenchmarkZeroBytes benchmarks memory zeroing at different sizes
func BenchmarkZeroBytes(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"16B", 16},
		{"64B", 64},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
		{"16KB", 16384},
		{"1MB", 1024 * 1024},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			data := make([]byte, sz.size)
			for i := range data {
				data[i] = 0xFF
			}

			b.ResetTimer()
			b.SetBytes(int64(sz.size))
			for i := 0; i < b.N; i++ {
				ZeroBytes(data)
			}
		})
	}
}

// BenchmarkZeroBytes_Parallel benchmarks parallel zeroing
func BenchmarkZeroBytes_Parallel(b *testing.B) {
	data := make([]byte, 1024)

	b.RunParallel(func(pb *testing.PB) {
		localData := make([]byte, len(data))
		copy(localData, data)

		for pb.Next() {
			ZeroBytes(localData)
		}
	})
}
