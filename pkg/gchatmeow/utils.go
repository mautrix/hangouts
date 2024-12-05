package gchatmeow

func GetPointer[T any](c T) *T {
	return &c
}
