package nn

import (
	"sync"
	"sync/atomic"
	"time"
)

// Server gathers evaluation requests from many concurrent games and runs them
// as single large network calls.
//
// It is not a server in the network sense - no sockets, no second process. One
// goroutine owns the session; every game worker hands over its position and
// blocks until its row of the answer comes back.
//
// This exists for the GPU, where nearly all the cost of a call is fixed
// overhead, so 256 positions in one call cost about what 4 do. On CPU the cost
// is flat per position at every batch size - measured 1.17ms at batch 1 and
// 1.22ms at batch 64 - so batching there buys nothing and Session should be
// used directly.
//
// The batch comes from concurrency across *games*, not within a search. Each
// worker plays its game sequentially and has exactly one evaluation outstanding
// at a time, so batch size is simply how many games are in flight. That is why
// no virtual loss and no change to the search were needed: run more games.
type Server struct {
	requests chan request
	session  Runner
	maxRows  int
	linger   time.Duration

	wg    sync.WaitGroup
	calls atomic.Int64
	rows  atomic.Int64
}

type request struct {
	in    []float32
	rows  int
	reply chan response
}

type response struct {
	policy []float32
	value  []float32
	err    error
}

// NewServer starts the batching goroutine. maxRows caps one call; queue is the
// request buffer, which wants to be at least the number of workers so a worker
// never blocks on handing over; linger is how long to keep gathering after the
// first request arrives.
//
// linger matters more than it looks. Without it the server fires as soon as
// anything is waiting, and since workers need CPU time between evaluations -
// descending the tree, applying moves, encoding - the queue is nearly empty
// immediately after a batch returns. The result is a full GPU call carrying two
// positions. A few hundred microseconds of waiting costs nothing against a call
// that takes twenty milliseconds, and fills it.
func NewServer(session Runner, maxRows, queue int, linger time.Duration) *Server {
	s := &Server{
		requests: make(chan request, queue),
		session:  session,
		maxRows:  maxRows,
		linger:   linger,
	}
	s.wg.Add(1)
	go s.loop()
	return s
}

// Close waits for in-flight work to finish. Every client must be done first.
func (s *Server) Close() {
	close(s.requests)
	s.wg.Wait()
}

// Stats reports how many network calls were made and how many positions went
// through them - their ratio is the average batch, which is the number worth
// watching to know the server is doing anything.
func (s *Server) Stats() (calls, rows int64) {
	return s.calls.Load(), s.rows.Load()
}

func (s *Server) loop() {
	defer s.wg.Done()

	var (
		buf     []float32
		pending []request
	)

	// One timer, reused: a fresh one per batch would allocate thousands of times
	// a second.
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	if !timer.Stop() {
		<-timer.C
	}

	for first := range s.requests {
		buf = append(buf[:0], first.in...)
		pending = append(pending[:0], first)
		rows := first.rows

		if s.linger > 0 {
			timer.Reset(s.linger)
		}
	gather:
		for rows < s.maxRows {
			if s.linger <= 0 {
				// Take what is already queued and go.
				select {
				case next, ok := <-s.requests:
					if !ok {
						break gather
					}
					buf = append(buf, next.in...)
					pending = append(pending, next)
					rows += next.rows
				default:
					break gather
				}
				continue
			}

			select {
			case next, ok := <-s.requests:
				if !ok {
					break gather
				}
				buf = append(buf, next.in...)
				pending = append(pending, next)
				rows += next.rows
			case <-timer.C:
				break gather
			}
		}
		if s.linger > 0 && !timer.Stop() {
			// Drain it only if it fired and nobody read it.
			select {
			case <-timer.C:
			default:
			}
		}

		policy, value, err := s.session.Run(buf, rows)
		s.calls.Add(1)
		s.rows.Add(int64(rows))

		// Hand each caller back its own rows.
		offset := 0
		for _, req := range pending {
			if err != nil {
				req.reply <- response{err: err}
				continue
			}
			req.reply <- response{
				policy: policy[offset*MoveCount : (offset+req.rows)*MoveCount],
				value:  value[offset : offset+req.rows],
			}
			offset += req.rows
		}
	}
}

// Client is one worker's handle on the server. Like Evaluator it is not safe
// for concurrent use - give every worker its own.
type Client struct {
	server *Server
	reply  chan response // reused across calls, so evaluating allocates nothing
}

func (s *Server) Client() *Client {
	return &Client{server: s, reply: make(chan response, 1)}
}

// Run submits a batch and waits for it, satisfying Runner so an Evaluator
// cannot tell the difference between this and a direct session.
func (c *Client) Run(in []float32, batch int) ([]float32, []float32, error) {
	c.server.requests <- request{in: in, rows: batch, reply: c.reply}
	result := <-c.reply
	return result.policy, result.value, result.err
}
