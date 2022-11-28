package generic

func GetMock[T any](mocks []any) T {
	for _, mock := range mocks {
		if res, ok := mock.(T); ok {
			return res
		}
	}

	panic("mock not found")
}

func GetService[T any](serviceCtx map[string]any) T {
	for _, object := range serviceCtx {
		if res, ok := object.(T); ok {
			return res
		}
	}

	panic("service not found")
}

type FilterFunc[T any] func(T) (bool, error)

func FilterSlice[T any](slice []T, filterFn FilterFunc[T]) ([]T, error) {
	var res []T

	for _, item := range slice {
		ok, err := filterFn(item)

		if err != nil {
			return nil, err
		}

		if ok {
			res = append(res, item)
		}
	}

	return res, nil
}
