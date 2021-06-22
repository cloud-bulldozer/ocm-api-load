package providers

type RoundRobinProvider struct {
	Items []interface{}
	Index int
}

func (r *RoundRobinProvider) GetItem() interface{} {
	if r.Index < len(r.Items) {
		p := r.Items[r.Index]
		if r.Index == len(r.Items)-1 {
			r.Index = 0
		} else {
			r.Index += 1
		}
		return p
	}
	return r.Items[r.Index]
}

func NewRoundRobinProvider(items ...interface{}) *RoundRobinProvider {
	return &RoundRobinProvider{
		Items: items,
		Index: 0,
	}
}
