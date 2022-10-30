package generic

import (
	"runtime"
	"sync"

	"go.uber.org/multierr"
)

func ParallelRangeMap[K comparable, V any](m map[K]V, fn func(k K, v V) error) error {
	var (
		wg   sync.WaitGroup
		errs = make([]error, len(m))
		sem  = make(chan struct{}, runtime.NumCPU())
	)

	idx := 0
	for k, v := range m {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int, k K, v V) {
			errs[idx] = fn(k, v)
			wg.Done()
			<-sem
		}(idx, k, v)
		idx++
	}

	wg.Wait()
	return multierr.Combine(errs...)
}
