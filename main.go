package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/pborman/getopt/v2"
	"io/ioutil"
	"load-generator/conf"
	"load-generator/models"
	"load-generator/utils"
	"log"
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
	expGenerator := utils.NewExponentialDistribution(rng, conf.ExpLambda)
	wg := sync.WaitGroup{}
	nreq := uint64(0)
	for {
		wg.Add(1)
		go launchVideo(nreq, zipfGenerator.Uint64(), videoList, &wg)
		nreq++
		secondsToWait := expGenerator.ExpFloat64()
		log.Println("Waiting for", secondsToWait, "seconds")
		time.Sleep(time.Duration(secondsToWait*1e6)*time.Microsecond + time.Hour) // TODO remove hour
	}
	wg.Wait() //nolint:govet
}

func parseArgs() {
	dryMode2 := getopt.BoolLong("dry-run", 't', "Launch the client in dry run mode (no actual video is retrieved)")
	getopt.Parse()
	dryMode = *dryMode2
}

func launchVideo(nreq uint64, u uint64, list []models.VideoMetadata, wg *sync.WaitGroup) {
	log.Printf("[#%d] Reproducing video n. %d => %s", nreq, u, list[u].Id)
	time.Sleep(time.Second * 2)

	/*opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
	)*/

	/*allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()*/

	// also set up a custom logger
	//taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	taskCtx, cancel := BuildHeadlessLocalChromeContext(900, true)
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(getVideoUrl(list[u])),
		/*chromedp.WaitVisible(`body > .footer-area`),
		chromedp.Click(`#iconPlayPause`, chromedp.NodeVisible),*/
	); err != nil {
		panic(err)
	}

	waitForEndOfVideo(&taskCtx, nreq)
	log.Printf("[#%d] End video n. %d => %s", nreq, u, list[u])
	wg.Done()
}

func waitForEndOfVideo(ctx *context.Context, nreq uint64) {
	for {
		log.Printf("Checking browser for req. no. %d", nreq)
		x := false
		str := ""
		if err := chromedp.Run(*ctx,
			chromedp.Evaluate(`window.metricsPushed === true`, &x)); err != nil {
			log.Println(err)
		}
		if err := chromedp.Run(*ctx,
			chromedp.Evaluate(`JSON.stringify(angular.element($('[ng-controller=DashController]')).scope().metricsArray)`, &str)); err != nil {
			log.Println(err)
		}
		log.Println(str)
		if x {
			log.Printf("Closing browser of req. no. %d...\n", nreq)
			if err := chromedp.Cancel(*ctx); err != nil {
				log.Println(err)
			}
		}
		time.Sleep(time.Minute)
	}
}

func getVideoUrl(v models.VideoMetadata) string {
	// TODO put in conf apiUrl and feUrl
	apiUrl := "http://video-metrics-collector.zion.alessandrodistefano.eu:8080/v1/video-reproduction"
	mpdUrl := fmt.Sprintf("%s/vms/videos/%s", conf.ServiceUrl, v.Id)
	// mpdUrl := "https://dash.akamaized.net/akamai/bbb_30fps/bbb_30fps.mpd"
	feUrl := "http://mora.zion.alessandrodistefano.eu:8880/samples/dash-if-reference-player-api-metrics-push/index.html"
	url := fmt.Sprintf("%s?url=%s&autoplay=true&apiUrl=%s", feUrl, mpdUrl, apiUrl)
	log.Println("Navigating to", url)
	return url
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
	log.Println(videoMetadata)
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
func NewLocalContext(timeout int, debug bool, opts []chromedp.ExecAllocatorOption) (context.Context, context.CancelFunc) {
	c, cancel_timeout := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	ac, cancel_alloc := chromedp.NewExecAllocator(c, opts...)

	var contextOpts = []chromedp.ContextOption{}
	if debug {
		contextOpts = []chromedp.ContextOption{
			chromedp.WithLogf(log.Printf),
			// chromedp.WithDebugf(log.Printf),
			chromedp.WithErrorf(log.Printf),
		}
	}

	ctx, cancel_context := chromedp.NewContext(ac, contextOpts...)

	return ctx, func() {
		cancel_timeout()
		cancel_alloc()
		cancel_context()
	}
}

func LocalChromeDefaults() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.WindowSize(3000, 2000),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-dev-shm-usage", "true"),
		chromedp.Flag("remote-debugging-address", "0.0.0.0"),
		chromedp.Flag("remote-debugging-port", "9222"), // TODO port based on nreq
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
	}
}

// Construct a headless chrome context using a locally-installed chrome instance
func BuildHeadlessLocalChromeContext(timeout int, debug bool) (context.Context, func()) {
	return NewLocalContext(timeout, debug, append(LocalChromeDefaults(),
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	))
}
