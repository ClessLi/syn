package reslove

type Queue []interface{}

func (q *Queue) Add(n interface{}) {
	*q = append(*q, n)
}

func (q *Queue) Poll() interface{} {
	p := (*q)[0]
	*q = (*q)[1:]
	return p
}

func (q Queue) Size() int {
	return len(q)
}

func (q Queue) IsEmpty() bool {
	return q.Size() == 0
}

func NewQueue() Queue {
	return Queue{}
}
