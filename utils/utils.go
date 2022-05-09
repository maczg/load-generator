package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"load-generator/resource"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

func HandleError() {
	if err := recover(); err != nil {
		log.Errorln(err)
		debug.PrintStack()
	}
}

func SetupCloseHandler(st *time.Time) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Infof("Sim duration: %s", time.Since(*st))
		os.Exit(0)
	}()
}

//GetRandomDurationBetween return random duration between max and min. min must be at least segment size duration in seconds
func GetRandomDurationBetween(min, max int) int {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Seed(time.Now().UnixNano())
	return (min + rng.Intn(max-min+1)) * 1e3
}

func GetVideoUrl(url string, v resource.VideoMetadata) (string, string) {
	originalUrl := fmt.Sprintf("%s/vms/videos/%s", url, v.Id)
	videofilesUrl := fmt.Sprintf("%s/videofiles/%s/video.mpd", url, v.Id)
	return originalUrl, videofilesUrl
}

type Counter struct {
	mu      sync.Mutex
	request map[string]int
}

func NewCounter() *Counter {
	return &Counter{request: make(map[string]int)}
}

// Inc increments the counter for the given key.
// keys: success, error, aborted,total,active
func (c *Counter) Inc(key string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.request[key]++
	c.mu.Unlock()
}

// Dec decrements the counter for the given key.
// keys: success, error, sum, active
func (c *Counter) Dec(key string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.request[key]--
	c.mu.Unlock()
}

// AddTo increments counter of given key for value times
//keys: success, error, sum, active
func (c *Counter) AddTo(key string, value int) {
	c.mu.Lock()
	c.request[key] += value
	c.mu.Unlock()
}

// GetRequestWith returns the current value of the counter for the given key.
func (c *Counter) GetRequestWith(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.request[key]
}
