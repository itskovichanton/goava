package async

import "context"

type Future interface {
	Await() interface{}
}

type future struct {
	await func(ctx context.Context) interface{}
}

func (f future) Await() interface{} {
	return f.await(context.Background())
}

type FutureResult struct {
	Result interface{}
	Err    error
}

func Exec(f func(interface{}) (interface{}, error), arg interface{}) Future {

	var fr FutureResult
	c := make(chan struct{})
	go func() {
		defer close(c)
		r, err := f(arg)
		fr = FutureResult{
			Result: r,
			Err:    err,
		}
	}()
	return future{
		await: func(ctx context.Context) interface{} {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				return fr
			}
		},
	}
}
