package inmemkv

import (
    "crypto/md5"
    "fmt"
    "io"
    "math/rand"
    "runtime"
    "strconv"
    "sync/atomic"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func BenchmarkInMemKeyValueParallel(b *testing.B) {
    b.SetParallelism(runtime.GOMAXPROCS(0))

    const numKeys = 30000
    const numKeysToRead = 10000

    inst := NewCache()

    for i := 0; i < numKeys; i++ {
        key := testComposeKey(i)
        inst.Set(key, strconv.Itoa(i*357))
    }

    read := make([]string, 0, numKeysToRead)
    for i := 0; i < numKeysToRead; i++ {
        key := testComposeKey(i)
        read = append(read, key)
    }

    i := atomic.Int64{}
    i.Store(0)

    b.ResetTimer()
    b.ReportAllocs()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            v := i.Load()
            i.Swap(v + 1)
            offset := v % numKeysToRead
            key := read[offset]
            inst.Get(key)
            if v%10 == 0 {
                inst.Delete(key)
            }
        }
    })
}

func BenchmarkInMemKeyValue(b *testing.B) {
    b.ReportAllocs()

    const numKeys = 30000

    inst := NewCache()

    for i := 0; i < numKeys; i++ {
        key := strconv.Itoa(i)
        inst.Set(key, strconv.Itoa(i*357))
    }
    key := strconv.Itoa(int(rand.Int63n(numKeys)))

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inst.Get(key)
    }
}

func testComposeKey(hint int) string {
    h := md5.New()
    _, _ = io.WriteString(h, strconv.Itoa(hint))
    return fmt.Sprintf("%x", h.Sum(nil))
}

func BenchmarkComposeKey(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        testComposeKey(i)
    }
}

func TestInMemKeyValueNonexpirable(t *testing.T) {
    is := assert.New(t)

    inst := NewCache()

    inst.Set("key1", "val1")
    inst.Set("key2", "val2")
    inst.Set("key3", "val3")

    raw, ok := inst.Get("key1")
    s := raw.(string)
    is.True(ok)
    is.NotNil(raw)
    is.Equal("val1", s)

    raw, ok = inst.Get("key2")
    s = raw.(string)
    is.True(ok)
    is.NotNil(raw)
    is.Equal("val2", s)

    raw, ok = inst.Get("key3")
    s = raw.(string)
    is.True(ok)
    is.NotNil(raw)
    is.Equal("val3", s)

    raw, ok = inst.Get("nonexistent")
    is.False(ok)
    is.Nil(raw)
}

func TestInMemKeyValueExpirable(t *testing.T) {
    is := assert.New(t)

    ttl := 10 * time.Millisecond

    inst := NewCache(WithTTL(ttl))

    inst.Set("key1", "val1")
    inst.Set("key2", "val2")
    inst.Set("key3", "val3")

    raw, ok := inst.Get("key1")
    s := raw.(string)
    is.True(ok)
    is.NotNil(raw)
    is.Equal("val1", s)

    raw, ok = inst.Get("key2")
    s = raw.(string)
    is.True(ok)
    is.NotNil(raw)
    is.Equal("val2", s)

    raw, ok = inst.Get("key3")
    s = raw.(string)
    is.True(ok)
    is.NotNil(raw)
    is.Equal("val3", s)

    raw, ok = inst.Get("nonexistent")
    is.False(ok)
    is.Nil(raw)

    time.Sleep(ttl)
    raw, ok = inst.Get("key1")
    is.False(ok)
    is.Nil(raw)

    raw, ok = inst.Get("key2")
    is.False(ok)
    is.Nil(raw)

    raw, ok = inst.Get("key3")
    is.False(ok)
    is.Nil(raw)

    raw, ok = inst.Get("nonexistent")
    is.False(ok)
    is.Nil(raw)

}
