package heap

type Heap struct {
	data []int
	less func(a, b int) bool
}

func (h *Heap) parent(i int) int {
	return (i - 1) / 2
}

func (h *Heap) left(i int) int {
	return 2*i + 1
}

func (h *Heap) right(i int) int {
	return 2*i + 2
}

func (h *Heap) swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

func (h *Heap) Insert(val int) {
	h.data = append(h.data, val)

	i := len(h.data) - 1

	for i > 0 && h.less(h.data[i], h.data[h.parent(i)]) {
		p := h.parent(i)
		h.swap(i, p)
		i = p
	}
}

func (h *Heap) heapifyDown(i int) {
	for {
		best := i
		left := h.left(i)
		right := h.right(i)

		if left < len(h.data) && h.less(h.data[left], h.data[best]) {
			best = left
		}

		if right < len(h.data) && h.less(h.data[right], h.data[best]) {
			best = right
		}

		if best == i {
			break
		}

		h.swap(i, best)
		i = best
	}
}

func (h *Heap) Pop() int {
	if len(h.data) == 0 {
		panic("heap empty")
	}

	root := h.data[0]
	last := h.data[len(h.data)-1]

	h.data[0] = last
	h.data = h.data[:len(h.data)-1]

	if len(h.data) > 0 {
		h.heapifyDown(0)
	}

	return root
}

func (h *Heap) Peek() int {
	return h.data[0]
}

func (h *Heap) Size() int {
	return len(h.data)
}

type MinHeap struct {
	Heap
}

func NewMinHeap() *MinHeap {
	return &MinHeap{
		Heap{
			data: []int{},
			less: func(a, b int) bool {
				return a < b
			},
		},
	}
}

type MaxHeap struct {
	Heap
}

func NewMaxHeap() *MaxHeap {
	return &MaxHeap{
		Heap{
			data: []int{},
			less: func(a, b int) bool {
				return a > b
			},
		},
	}
}
