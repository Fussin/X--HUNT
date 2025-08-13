package syncx

type Semaphore struct{ ch chan struct{} }

func NewSemaphore(n int) *Semaphore { return &Semaphore{ch: make(chan struct{}, n)} }

func (s *Semaphore) Acquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() { select { case <-s.ch: default: } }
func (s *Semaphore) Size() int { return len(s.ch) }
func (s *Semaphore) Cap() int  { return cap(s.ch) }
