package workqueue

type RateLimitingInterface interface {
	DelayingInterface

	AddRateLimited(item interface{})

	Forget(item interface{})

	NumRequeues(item interface{}) int
}

func NewRateLimitingQueue(rateLimiter RateLimiter) RateLimitingInterface {
	return &rateLimitingType{
		DelayingInterface: NewDelayingQueue(),
		rateLimiter:       rateLimiter,
	}
}

func NewNamedRateLimitingQueue(rateLimiter RateLimiter, name string) RateLimitingInterface {
	return &rateLimitingType{
		DelayingInterface: NewNamedDelayingQueue(name),
		rateLimiter:       rateLimiter,
	}
}

type rateLimitingType struct {
	DelayingInterface

	rateLimiter RateLimiter
}

func (q *rateLimitingType) AddRateLimited(item interface{}) {
	q.DelayingInterface.AddAfter(item, q.rateLimiter.When(item))
}

func (q *rateLimitingType) NumRequeues(item interface{}) int {
	return q.rateLimiter.NumRequeues(item)
}

func (q *rateLimitingType) Forget(item interface{}) {
	q.rateLimiter.Forget(item)
}
