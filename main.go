package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"load-generator/player"
	"load-generator/resource"
	"load-generator/utils"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	dryMode             = kingpin.Flag("dry-run", "dry-run mode").Default("false").Short('D').Envar("DRYRUN").Bool()
	maxReq              = kingpin.Flag("maxreq", "max concurrent goroutine").Default("10").Short('m').Envar("MAX_REQUESTS").Int()
	serviceUrl          = kingpin.Flag("service", "target url of load generation").Default("http://cloud.gollo1.particles.dieei.unict.it").Short('s').Envar("SERVICE_URL").String()
	zipfS               = kingpin.Flag("zipfs", "param S of zipf func").Default("1.01").Short('S').Envar("ZIPFS").Float()
	zipfV               = kingpin.Flag("zipfv", "param V of zipf func").Default("1").Short('V').Envar("ZIPFV").Float()
	expLambda           = kingpin.Flag("explambda", "param exponentialm distribution").Default("0.1").Short('e').Envar("EXP_LAMBDA").Float()
	simDuration         = kingpin.Flag("duration", "duration of simulation").Default("30m").Short('d').Envar("SIM_DURATION").Duration()
	lc                  = kingpin.Flag("lc", "load-cure [00|01|02|03]").Default("00").Short('l').String()
	variant             = kingpin.Flag("variant", "variant [CO|ECO1|ECO2]").Default("CO").Short('v').String()
	experiment          = strconv.Itoa(int(time.Now().UnixNano()))
	startTimeSimulation = time.Now()
)

func init() {
	prometheus.Register(player.TotalByteRcv)
	prometheus.Register(player.Req)
}

func main() {
	kingpin.Parse()
	info := parseExpInfo()
	videoList := getVideoSlice()
	utils.SetupCloseHandler(&startTimeSimulation)

	N := uint64(len(videoList))
	rng := rand.New(rand.NewSource(0))
	zipfGenerator := rand.NewZipf(rng, *zipfS, *zipfV, N-1)
	expGenerator := utils.NewExponentialDistribution(rng, *expLambda)

	requestSem := make(chan struct{}, *maxReq)
	wg := sync.WaitGroup{}
	nreq := uint64(0)

	wg.Add(1)
	go setupPromServer()

	for {
		//test. Lock here in order of not waste for cicle ?? unlock inside goroutine
		requestSem <- struct{}{}
		wg.Add(1)
		go player.Play(info, nreq, zipfGenerator.Uint64(), videoList, &wg, requestSem)
		secondsToWait := expGenerator.ExpFloat64()
		time.Sleep(time.Duration(secondsToWait * 1e6))
		nreq++
	}

	wg.Wait() //nolint:govet

}

func parseExpInfo() resource.ExperimentInfo {
	return resource.ExperimentInfo{
		Experiment: experiment,
		Variant:    *variant,
		LoadCurve:  *lc,
		DryRun:     *dryMode,
		ServiceUrl: *serviceUrl,
	}
}

func getVideoSlice() (videoMetadata []resource.VideoMetadata) {
	resp, err := http.Get(fmt.Sprintf("%s/vms/videos", *serviceUrl))
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
	return
}

func setupPromServer() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
