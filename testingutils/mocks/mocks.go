package mocks

func GetMock[T any](mocks []any) T {
	for _, mock := range mocks {
		if res, ok := mock.(T); ok {
			return res
		}
	}

	panic("mock not found")
}
