package load_balance

type LoadBalance interface {
	Add(...string) error
	Get(string) (string, error)
	Update()
}

type LbType int

const (
	LbRandom LbType = iota
	LbRoundRobin
	LbWeightRoundRobin
	LbConsistentHash
)

func LoadBalanceFactory(lbType LbType) *LoadBalance {
	switch lbType {
	case LbRandom:
	case LbRoundRobin:
	case LbWeightRoundRobin:
	case LbConsistentHash:
	default:
	}
	return nil
}
