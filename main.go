package main

import (
	"encoding/json"
	"fmt"
	metrics "github.com/massimo-gollo/DASHpher/models"
	_ "github.com/massimo-gollo/godash/player"
	"github.com/pborman/getopt/v2"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"load-generator/conf"
	"load-generator/models"
	"load-generator/player"
	"load-generator/utils"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var dryMode bool
var maxReq int
var targeturl string

var counter *utils.Counter

func main() {

	counter = utils.NewCounter()
	//prepare counter - just to know counter values
	//success - complete without error in segment
	counter.AddTo("success", 0)
	//aborted - some reason not downloaded all segment
	counter.AddTo("aborted", 0)
	//completed but dropped some segment
	counter.AddTo("witherror", 0)
	//total request made
	counter.AddTo("total", 0)
	//total request in queue
	counter.AddTo("active", 0)
	counter.AddTo("stalls", 0)

	parseArgs()
	videoList := getVideoSlice()
	N := uint64(len(videoList))
	rng := rand.New(rand.NewSource(0))
	zipfGenerator := rand.NewZipf(rng, conf.ZipfS, conf.ZipfV, N-1)
	log.Infoln("Number of video:", N)
	log.Infof("Max concurrent request %d (default %d)", maxReq, 10)
	expGenerator := utils.NewExponentialDistribution(rng, conf.ExpLambda)
	goroutineBuffer := make(chan struct{}, maxReq)
	wg := sync.WaitGroup{}
	nreq := uint64(0)
	requestsMetrics := make(map[uint64]*metrics.ReproductionMetrics)
	log.Infof("start in %d second", 5)
	time.Sleep(5 * time.Second)

	stSim := time.Now()
	SetupCloseHandler(&stSim)
	log.Infof("Start simulation at %s - target url %s", stSim.Format("15:04:05"), targeturl)

	go UpdateStatus(&nreq, &stSim)

	for {
		//test. Lock here in order of not waste for cicle ?? unlock inside goroutine
		goroutineBuffer <- struct{}{}
		requestsMetrics[nreq] = metrics.NewReproductionMetrics()
		wg.Add(1)
		go player.Play(targeturl, counter, requestsMetrics[nreq], nreq, zipfGenerator.Uint64(), videoList, &wg, false, goroutineBuffer)
		nreq++
		secondsToWait := expGenerator.ExpFloat64()
		//log.Println("Waiting for", time.Duration(secondsToWait*1e6), "seconds")
		time.Sleep(time.Duration(secondsToWait * 1e6))
	}

	wg.Wait() //nolint:govet

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

func parseArgs() {
	dryMode2 := getopt.BoolLong("dry-run", 't', "Launch the client in dry run mode (no actual video is retrieved)")
	concurrent := getopt.IntLong("max-req", 'c', 10, "Specify max Number of concurrent request - (max goroutine in execution)")
	url := getopt.String('u', "http://cloud.gollo1.particles.dieei.unict.it", "target url")
	getopt.Parse()
	dryMode = *dryMode2
	maxReq = *concurrent
	targeturl = *url
}

func UpdateStatus(nreq *uint64, t *time.Time) {
	pollingTick := time.Tick(20 * time.Second)
	for {
		select {
		case <-pollingTick:
			total := counter.GetRequestWith("total")
			witherr := counter.GetRequestWith("witherror")
			aborter := counter.GetRequestWith("aborted")
			succeded := counter.GetRequestWith("success")
			active := counter.GetRequestWith("active")
			stalls := counter.GetRequestWith("stalls")
			log.Infof("init at: %s | duration: %s | Active: %d/%d | Success %d/%d | Finish with error: %d/%d | Aborted: %d/%d | Stalls %d",
				t.Format("15:04:05"),
				time.Since(*t).Truncate(time.Second).String(),
				active, maxReq,
				succeded, total,
				witherr, total,
				aborter, total, stalls)
		}
	}
}

func getVideoSlice() (videoMetadata []models.VideoMetadata) {
	resp, err := http.Get(fmt.Sprintf("%s/vms/videos", targeturl))
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
