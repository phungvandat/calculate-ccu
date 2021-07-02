package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	httpAddr = ":1234"
)

type hdl struct {
	rClient *redis.Client
	w       *worker
}

func (h *hdl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		data        interface{}
		err         error
		code        = http.StatusOK
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	)
	defer func() {
		cancel()
		mess := fmt.Sprintf("endpoint: %s\t", r.URL.Path)
		if err != nil {
			mess += fmt.Sprintf(" err: %v", err)
		}
		log.Println(mess)
	}()

	select {
	case <-ctx.Done():
		err = errors.New("timeout")
	case <-func() chan bool {
		resChn := make(chan bool)
		go func() {
			defer close(resChn)

			switch r.URL.Path {
			case "/ccu":
				data, err = h.currentCCU(ctx, r)
			default:
				err = h.calculate(r)
			}

			resChn <- true
		}()
		return resChn
	}():
	}

	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		data = map[string]interface{}{
			"err": err.Error(),
		}
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
	writerErr := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("write json failed: %v\n", writerErr)
	}
}

func (h *hdl) currentCCU(ctx context.Context, r *http.Request) (interface{}, error) {
	var currentTime = time.Now().Format("200601021504")
	cmd, err := h.rClient.PFCount(ctx, currentTime).Result()
	if err != nil {
		panic(err)
	}
	return map[string]interface{}{
		"num": cmd,
	}, nil
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func (h *hdl) calculate(r *http.Request) error {
	// mock handle to affect to resource
	_ = fibonacci(20)
	var currentTime = time.Now().Format("200601021504") // YYYYMMDDHHMM
	// random a number
	// see it like to the user ID
	var num = fmt.Sprintf("%v", rand.Intn(100000))
	h.w.Receive(&job{
		key: currentTime,
		el:  num,
	})
	return nil
}
