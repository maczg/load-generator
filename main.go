package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/chromedp"
	_ "github.com/massimo-gollo/godash/player"
	"github.com/pborman/getopt/v2"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"load-generator/conf"
	"load-generator/httpreq"
	"load-generator/models"
	"load-generator/player"
	"load-generator/utils"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var dryMode bool

func main() {
	parseArgs()
	videoList := getVideoSlice()
	N := uint64(len(videoList))
	rng := rand.New(rand.NewSource(0))
	zipfGenerator := rand.NewZipf(rng, conf.ZipfS, conf.ZipfV, N-1)
	log.Println("Number of video:", N)
	expGenerator := utils.NewExponentialDistribution(rng, conf.ExpLambda)
	maxNbConcurrentGoroutines := flag.Int("maxNbConcurrentGoroutines", 300, "the number of goroutines that are allowed to run concurrently")
	flag.Parse()
	concurrentGoroutines := make(chan struct{}, *maxNbConcurrentGoroutines)
	wg := sync.WaitGroup{}
	_ = initPortsChan()
	nreq := uint64(0)
	for {
		wg.Add(1)
		go player.Reproduction(nreq, zipfGenerator.Uint64(), videoList, &wg, false, concurrentGoroutines)
		nreq++
		secondsToWait := expGenerator.ExpFloat64()
		log.Println("Waiting for", secondsToWait, "seconds")
		//time.Sleep(time.Duration(secondsToWait*1e6) * time.Microsecond) // TODO remove hour

		time.Sleep(time.Millisecond * 100)
		//time.Sleep(time.Duration(secondsToWait*1e6) * time.Microsecond)
	}
	wg.Wait() //nolint:govet
}

func initPortsChan() chan int {
	ch := make(chan int, conf.MaxExposedPorts)
	for i := 0; int64(i) < conf.MaxExposedPorts; i++ {
		ch <- 9222 + i
	}
	return ch
}

func parseArgs() {
	dryMode2 := getopt.BoolLong("dry-run", 't', "Launch the client in dry run mode (no actual video is retrieved)")
	getopt.Parse()
	dryMode = *dryMode2
}

func launchVideo(nreq uint64, u uint64, list []models.VideoMetadata, wg *sync.WaitGroup, portsChan chan int) {
	log.Printf("[Req#%d] Reproducing video n. %d => %s", nreq, u, list[u].Id)
	if dryMode {
		time.Sleep(time.Second * 2)

	} else {
		var port int
		select {
		case port = <-portsChan:
			log.Printf("[Req#%d] Acquiring port %d", nreq, port)
		default:
			port = 0
		}

		taskCtx, cancel := BuildHeadlessLocalChromeContext(conf.MaxExecutionTime, true, port)
		defer cancel()
		t, _ := httpreq.GetVideoUrl(list[u])
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(t),
		); err != nil {
			panic(err)
		}
		waitForEndOfVideo(&taskCtx, nreq)
		if port > 0 {
			log.Printf("[Req#%d] Releasing port %d", nreq, port)
			portsChan <- port
		}
	}
	log.Printf("[Req#%d] End video n. %d => %s", nreq, u, list[u])
	wg.Done()
}

func waitForEndOfVideo(ctx *context.Context, nreq uint64) {
	pollingTick := time.Tick(10 * time.Second)
	for {
		select {
		case <-pollingTick:
			log.Printf("[Req#%d] Checking browser", nreq)
			x := false
			var metricsArray []interface{}
			if err := chromedp.Run(*ctx,
				chromedp.Evaluate(`window.metricsPushed === true`, &x)); err != nil {
				log.Println(err)
			}
			if err := chromedp.Run(*ctx,
				chromedp.Evaluate(`angular.element($('[ng-controller=DashController]')).scope().metricsArray`,

					&metricsArray)); err != nil {
				log.Println(err)
			}

			log.Printf("[Req#%d] Checking browser len(metricsArray) = %d", nreq, len(metricsArray))
			if x {
				log.Printf("[Req#%d]Closing browser...\n", nreq)
				if err := chromedp.Cancel(*ctx); err != nil {
					log.Println(err)
				}
				return
			}
		case <-(*ctx).Done():
			log.Printf("[Req#%d] Timeout for metrics push exceeded", nreq)
			return
		}
	}
}

func getVideoSlice() (videoMetadata []models.VideoMetadata) {
	resp, err := http.Get(fmt.Sprintf("%s/vms/videos", conf.ServiceUrl))
	if err != nil {
		log.Fatal("Unable to get videos from the server: ", err)
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to read request body", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("fetching information from the vms microservice failed with status code %d and body %s",
			resp.StatusCode, resp.Body)
	}
	err = json.Unmarshal(body, &videoMetadata)
	if err != nil {
		log.Fatal("Unable to unmarshal json array", err)
	}
	writeToFile(videoMetadata)
	return
}

func writeToFile(metadata []models.VideoMetadata) {
	f, err := os.Create("/tmp/dat2")
	check(err)
	defer f.Close()
	for _, m := range metadata {
		_, err = f.WriteString(fmt.Sprintf("%s\n", m.Id))
		check(err)
	}
	check(f.Sync())
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Create and return a new local context. Called by more specific constructors.
func NewLocalContext(timeout int64, debug bool, opts []chromedp.ExecAllocatorOption) (context.Context, context.CancelFunc) {
	c, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	ac, cancelAlloc := chromedp.NewExecAllocator(c, opts...)

	var contextOpts = []chromedp.ContextOption{}
	if debug {
		contextOpts = []chromedp.ContextOption{
			chromedp.WithLogf(log.Printf),
			// chromedp.WithDebugf(log.Printf),
			chromedp.WithErrorf(log.Printf),
		}
	}

	ctx, cancelContext := chromedp.NewContext(ac, contextOpts...)

	return ctx, func() {
		cancel()
		cancelAlloc()
		cancelContext()
	}
}

func LocalChromeDefaults() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.WindowSize(3000, 2000),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-dev-shm-usage", "true"),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
	}
}

// Construct a headless chrome context using a locally-installed chrome instance
func BuildHeadlessLocalChromeContext(timeout int64, debug bool, port int) (context.Context, func()) {
	args := append(LocalChromeDefaults(),
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)

	if port > 0 {
		args = append(args, chromedp.Flag("remote-debugging-address", "0.0.0.0"),
			chromedp.Flag("remote-debugging-port", fmt.Sprintf("%d", port)))
	}

	return NewLocalContext(timeout, debug, args)
}
