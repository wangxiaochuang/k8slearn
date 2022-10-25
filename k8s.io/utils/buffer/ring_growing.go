package buffer

type RingGrowing struct {
	data     []interface{}
	n        int // data的长度
	beg      int // 当前位置
	readable int // 可读的数量，为0不可读
}

func NewRingGrowing(initialSize int) *RingGrowing {
	return &RingGrowing{
		data: make([]interface{}, initialSize),
		n:    initialSize,
	}
}

func (r *RingGrowing) ReadOne() (data interface{}, ok bool) {
	if r.readable == 0 {
		return nil, false
	}
	r.readable--
	element := r.data[r.beg]
	// 读了后就设置为nil，好回收
	r.data[r.beg] = nil
	// 读到最后一个元素了，所以从头开始
	if r.beg == r.n-1 {
		r.beg = 0
	} else {
		// 下一个元素位置
		r.beg++
	}
	return element, true
}

func (r *RingGrowing) WriteOne(data interface{}) {
	// 目前的数量等于容量了，需要扩容
	if r.readable == r.n {
		newN := r.n * 2
		newData := make([]interface{}, newN)
		to := r.beg + r.readable
		// 迁移数据
		if to <= r.n {
			// copy(newData, r.data[:])
			copy(newData, r.data[r.beg:to])
		} else {
			copied := copy(newData, r.data[r.beg:])
			copy(newData[copied:], r.data[:(to%r.n)])
		}
		r.beg = 0
		r.data = newData
		r.n = newN
	}
	r.data[(r.readable+r.beg)%r.n] = data
	r.readable++
}
