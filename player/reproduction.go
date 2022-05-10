package player

import (
	metrics "github.com/massimo-gollo/DASHpher/models"
	"github.com/massimo-gollo/DASHpher/reproduction"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"load-generator/resource"
	"load-generator/utils"
	"strconv"
	"sync"
	"time"
)

var (
	Req = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_reproduction",
		Help: "total_reproduction_with_status",
	}, []string{"experiment", "loadcurve", "variant", "status"})

	TotalByteRcv = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "total_byte_rcv",
		Help: "total byte rcv from target"},
		[]string{"experiment", "loadcurve", "variant"})

	ResponseTime = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "response_time_request",
		Help: "response_time_of_request"},
		[]string{"experiment", "loadcurve", "variant", "request"})
)

func Play(expInfo resource.ExperimentInfo, nr uint64, u uint64, list []resource.VideoMetadata, w *sync.WaitGroup, sem chan struct{}) {
	defer utils.HandleError()
	defer w.Done()
	defer log.Printf("[Req#%d] End video n. %d => %s", nr, u, list[u].Id)
	log.Printf("[Req#%d] Reproducing video n. %d => %s", nr, u, list[u].Id)
	Req.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, "total").Inc()
	Req.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, "active").Inc()
	metric := metrics.NewReproductionMetrics()
	if expInfo.DryRun {
		time.Sleep(time.Second * 10)
	} else {
		original, _ := utils.GetVideoUrl(expInfo.ServiceUrl, list[u])
		metric.ContentUrl = original
		metric.ReproductionID = nr
		_ = reproduction.Stream(metric, "h264", "conventional", 1080, 240000, 2, 10, nr)
		switch metric.Status {
		case metrics.Aborted:
			Req.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, "aborted").Inc()
		case metrics.Error:
			ResponseTime.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, strconv.Itoa(int(nr))).Add(metric.FetchMpdInfo.RTT2FirstByte.Minutes())
			Req.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, "error").Inc()
			TotalByteRcv.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant).Add(float64(metric.TotalByteDownloaded))
		default:
			ResponseTime.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, strconv.Itoa(int(nr))).Add(metric.FetchMpdInfo.RTT2FirstByte.Minutes())
			TotalByteRcv.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant).Add(float64(metric.TotalByteDownloaded))
			Req.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, "success").Inc()
		}
	}
	Req.WithLabelValues(expInfo.Experiment, expInfo.LoadCurve, expInfo.Variant, "active").Dec()
	<-sem
}
