# Benchmark

## Rate Limiter Result
 
code:

```
func Test_DecrementPerformance(t *testing.T) {
	startTime := time.Now()
	var batchQuotaSize int64 = 10000
	maxBucketQuotaSize := batchQuotaSize
	redisCleanTimeout := time.Second * 10
	limiter, err := NewRateLimiter(batchQuotaSize, maxBucketQuotaSize, redisCleanTimeout, pool)
	assert.NoError(t, err)

	// case: first time have two requests concurrently
	c := context.Background()
	g, _ := errgroup.WithContext(c)
	workerSize := int(batchQuotaSize)
	for i := 0; i < workerSize; i++ {
		g.Go(func() error {
			ctx := goctx.Background()
			errRoutine := limiter.Decrement(ctx, "test", "performance", 1)
			return errRoutine
		})
	}
	err = g.Wait()
	assert.NoError(t, err)
	elapsed := time.Since(startTime)
	fmt.Println("elapsed time:", elapsed)
}
```
 
 
 
—— Running 100 goroutine concurrently ——

time spent:

round 1: 5.638164ms

round 2: 6.696378ms

round 3: 6.545043ms

round 4: 5.484312ms

round 5: 7.057415ms

advantage time spent: 6.28ms
 
 
 
—— Running 1000 goroutine concurrently ——

time spent:

round 1: 79.971651ms

round 2: 83.729121ms

round 3: 61.664404ms

round 4: 74.655635ms

round 5: 64.065122ms

advantage time spent: 72.82ms
 
 
 
—— Running 10000 goroutine concurrently ——

use local cache + mu.Lock(), time spent:

round 1: 711.632926ms

round 2: 813.007729ms

round 3: 736.550876ms

round 4: 627.674146ms

round 5: 698.717322ms

advantage time spent: 717.52ms
 

 
have local cache + single flight, time spent:
 
 round 1: 648.055927ms
 
 round 2: 661.128501ms
 
 round 3: 733.405936ms
 
 round 4: 634.598726ms
 
 round 5: 633.72743ms
 
 advantage time spent: 662.182ms
 
 

have single flight and not have to display logger, time spent:
 
 round 1: 42.51068ms
 
 round 2: 44.997799ms
 
 round 3: 43.314509ms
 
 round 4: 42.385853ms
 
 round 5: 39.547247ms
 
 advantage time spent: 42.55ms
