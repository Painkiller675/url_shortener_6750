package service

import "testing"

func BenchmarkService(b *testing.B) {
	str := "testString"
	b.Run("GetRandString_Func", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetRandString(str)
		}
	})

}
