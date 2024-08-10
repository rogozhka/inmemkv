package inmemkv

var _ KeyValueStorage = (*inMemKeyValue)(nil)

type KeyValueStorage interface {
    Set(key string, value any)
    Get(key string) (any, bool)
    Is(key string) bool
    Delete(key string)
    Reset()
}
