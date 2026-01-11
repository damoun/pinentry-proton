package protocol

import (
	"strings"
	"testing"
)

// BenchmarkPercentEncode_Sizes benchmarks percent encoding at different sizes
func BenchmarkPercentEncode_Sizes(b *testing.B) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Small_8B", []byte("password")},
		{"Medium_32B", []byte(strings.Repeat("a", 32))},
		{"Medium_128B", []byte(strings.Repeat("a", 128))},
		{"Large_1KB", []byte(strings.Repeat("a", 1024))},
		{"Large_4KB", []byte(strings.Repeat("a", 4096))},
		{"Large_16KB", []byte(strings.Repeat("a", 16384))},
		{"SpecialChars", []byte("p@ss!w#rd$%^&*()")},
		{"Unicode", []byte("пароль日本語🔐")},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(tt.data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = PercentEncode(tt.data)
			}
		})
	}
}

// BenchmarkPercentEncode_ManySpecialChars benchmarks worst case (many special chars)
func BenchmarkPercentEncode_ManySpecialChars(b *testing.B) {
	// Worst case: every character needs encoding (3x expansion)
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PercentEncode(data)
	}
}

// BenchmarkUnescapeArg benchmarks argument unescaping
func BenchmarkUnescapeArg(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"NoEscape", "hello world"},
		{"SingleEscape", "hello%20world"},
		{"MultipleEscapes", "hello%20world%0Atest%25"},
		{"AllEscaped", "%48%65%6C%6C%6F"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = UnescapeArg(tt.input)
			}
		})
	}
}

// BenchmarkEscapeArg benchmarks argument escaping
func BenchmarkEscapeArg(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"NoEscape", "hello"},
		{"WithSpaces", "hello world test"},
		{"WithNewlines", "hello\nworld\ntest"},
		{"WithSpecial", "test%+\n\r"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EscapeArg(tt.input)
			}
		})
	}
}

// BenchmarkPercentEncode_RoundTrip benchmarks encode + decode
func BenchmarkPercentEncode_RoundTrip(b *testing.B) {
	data := []byte("test password with special chars !@#$%")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded := PercentEncode(data)
		_ = UnescapeArg(encoded)
	}
}

// BenchmarkPercentEncode_Parallel benchmarks concurrent encoding
func BenchmarkPercentEncode_Parallel(b *testing.B) {
	data := []byte(strings.Repeat("test", 256))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = PercentEncode(data)
		}
	})
}

// BenchmarkSessionReset benchmarks session reset operations
func BenchmarkSessionReset(b *testing.B) {
	session := &Session{
		description: "test description",
		prompt:      "PIN:",
		title:       "Test Title",
		keyInfo:     "ABCD1234",
		error:       "error",
		sensitiveData: [][]byte{
			[]byte("password1"),
			[]byte("password2"),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.reset()
		// Repopulate for next iteration
		session.description = "test description"
		session.prompt = "PIN:"
		session.title = "Test Title"
		session.keyInfo = "ABCD1234"
		session.error = "error"
		session.sensitiveData = [][]byte{
			[]byte("password1"),
			[]byte("password2"),
		}
	}
}
