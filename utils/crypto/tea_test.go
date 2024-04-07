package crypto

import (
	"crypto/rand"
	"testing"
)

func TestTea(t *testing.T) {
	s := "hello tea"
	key := []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	encryp := QQTeaEncrypt([]byte(s), key)
	raw := string(QQTeaDecrypt(encryp, key))
	if s == raw {
		t.Log("pass")
	} else {
		t.Error("fatal")
	}
}
func BenchmarkTea(b *testing.B) {
	b.Run("16-16", func(b *testing.B) {
		data := getRandomBytes(16)
		key := getRandomBytes(16)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d := QQTeaEncrypt(data, key)
			_ = QQTeaDecrypt(d, key)
		}
	})
	b.Run("512-16", func(b *testing.B) {
		data := getRandomBytes(512)
		key := getRandomBytes(16)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d := QQTeaEncrypt(data, key)
			_ = QQTeaDecrypt(d, key)
		}
	})
	b.Run("1024-16", func(b *testing.B) {
		data := getRandomBytes(1024)
		key := getRandomBytes(16)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d := QQTeaEncrypt(data, key)
			_ = QQTeaDecrypt(d, key)
		}
	})
	b.Run("2048-16", func(b *testing.B) {
		data := getRandomBytes(2048)
		key := getRandomBytes(16)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d := QQTeaEncrypt(data, key)
			_ = QQTeaDecrypt(d, key)
		}
	})

}

func getRandomBytes(n int) (data []byte) {
	data = make([]byte, n)
	_, _ = rand.Read(data)
	return
}
