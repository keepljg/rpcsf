package swr

type smoothWeight struct {
	Item interface{}
	Weight int
	CurrentWeight int
}

func NewSw() *Sw{
	return &Sw{
		sms:make([]*smoothWeight, 0),
	}
}

type Sw struct {
	sms []*smoothWeight
	len int
	total int
}

func (s *Sw) Add(item interface{}, weight int) {
	sm := &smoothWeight{
		Item: item,
		Weight: weight,
	}
	s.sms = append(s.sms, sm)
	s.len += 1
	s.total += weight
	return
}

func (s *Sw) Get() interface{} {
	if s.len == 0 {
		return nil
	}
	var res *smoothWeight
	for i := 0; i < s.len; i ++ {
		s := s.sms[i]
		s.CurrentWeight += s.Weight
		if res == nil || res.CurrentWeight < s.Weight {
			res = s
		}
	}
	res.CurrentWeight -= s.total
	return res.Item
}
