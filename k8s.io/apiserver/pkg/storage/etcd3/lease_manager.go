package etcd3

const (
	defaultLeaseReuseDurationSeconds = 60
	defaultLeaseMaxObjectCount       = 1000
)

type LeaseManagerConfig struct {
	// ReuseDurationSeconds specifies time in seconds that each lease is reused
	ReuseDurationSeconds int64
	// MaxObjectCount specifies how many objects that a lease can attach
	MaxObjectCount int64
}

func NewDefaultLeaseManagerConfig() LeaseManagerConfig {
	return LeaseManagerConfig{
		ReuseDurationSeconds: defaultLeaseReuseDurationSeconds,
		MaxObjectCount:       defaultLeaseMaxObjectCount,
	}
}
