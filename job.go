package main

import (
	"context"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

const (
	maxBuff = 10000
)

type worker struct {
	pipeliner redis.Pipeliner
	jobChn    chan *job
	ctx       context.Context
	cancel    func()
	wg        sync.WaitGroup
}

type job struct {
	key string
	el  interface{}
}

func newWorker(rClient *redis.Client) (*worker, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &worker{
		pipeliner: rClient.Pipeline(),
		jobChn:    make(chan *job, maxBuff),
		ctx:       ctx,
		cancel:    cancel,
	}
	return w, w.Close
}

// Receive send job to channel
func (w *worker) Receive(j *job) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.jobChn <- j
	}()
}

// Do will handle write hyperloglog to redis with max batch is 10000
func (w *worker) Do() {
	log.Println("job running...")
	idx, done := 0, false

	for {
	ChildLoop:
		for {
			if idx == maxBuff {
				break
			}
			select {
			case job, ok := <-w.jobChn:
				if !ok {
					done = true
					break ChildLoop
				}
				w.pipeliner.PFAdd(context.Background(), job.key, job.el)
				idx++
			default:
				break ChildLoop
			}
		}
		if idx > 0 {
			_, err := w.pipeliner.Exec(context.Background())
			if err != nil {
				panic(err)
			}
			idx = 0
		}
		if done {
			break
		}
	}
	w.cancel()
	log.Println("job finish")
}

// Close achieve graceful shutdown
func (w *worker) Close() {
	w.wg.Wait()
	close(w.jobChn)
	<-w.ctx.Done()
}
