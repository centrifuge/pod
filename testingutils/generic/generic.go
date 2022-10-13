package generic

func GetMock[T any](mocks []any) T {
	for _, mock := range mocks {
		if res, ok := mock.(T); ok {
			return res
		}
	}

	panic("mock not found")
}

func GetObject[T any](serviceCtx map[string]any) T {
	for _, object := range serviceCtx {
		if res, ok := object.(T); ok {
			return res
		}
	}

	panic("object not found")
}
