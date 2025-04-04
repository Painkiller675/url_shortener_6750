package memory

import (
	"context"
	"testing"

	"github.com/Painkiller675/url_shortener_6750/internal/config"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
)

func BenchmarkMemory(b *testing.B) {
	l, _ := logger.NewZapLogger(config.StartOptions.LogLvl)
	memStor := NewStorage(l.Logger)
	b.Run("MemoryStore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			memStor.StoreAlURL(context.Background(), "al123", "url123", "123")
		}
	})

}
