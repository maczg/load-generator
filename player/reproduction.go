package player

import (
	"github.com/massimo-gollo/godash/P2Pconsul"
	glob "github.com/massimo-gollo/godash/global"
	"github.com/massimo-gollo/godash/http"
	"github.com/massimo-gollo/godash/player"
	log "github.com/sirupsen/logrus"
	"load-generator/httpreq"
	"load-generator/models"
	"load-generator/utils"
	"sync"
	"time"
)

func Reproduction(nreq uint64, u uint64, list []models.VideoMetadata, wg *sync.WaitGroup, dryMode bool, concurrentGoroutines chan struct{}) {
	defer utils.HandleError()

	defer log.Printf("[Req#%d] End video n. %d => %s", nreq, u, list[u])
	defer wg.Done()
	concurrentGoroutines <- struct{}{}
	log.Printf("[Req#%d] Reproducing video n. %d => %s", nreq, u, list[u].Id)

	if dryMode {
		time.Sleep(time.Second * 10)
	} else {
		originalUrl, directUrl := httpreq.GetVideoUrl(list[u])

		//url := "http://staging.massimogollo.it/videofiles/623c8a8008e7d25d8861139c/video.mpd"
		mpd, err := httpreq.ReadMPD(originalUrl)
		if err != nil {
			return
		}

		//structlist with only one mpd
		structList := []http.MPD{*mpd.Mpd}
		maxSegments, segmentDurationArray := http.GetSegmentDetails(structList, 0)
		segmentDuration := segmentDurationArray[0]
		lastSegmentDuration := http.SplitMPDSegmentDuration(structList[0].MaxSegmentDuration)
		mpdStreamDuration := segmentDuration*(maxSegments-1) + lastSegmentDuration

		var printHeadersData map[string]string

		var Noden = P2Pconsul.NodeUrl{}
		//TODO codec hardcoded h264
		player.MainStream(structList, glob.DebugFile, false, "h264", glob.CodecName, 2160,
			mpdStreamDuration*1000, 30, 2, "conventional", directUrl,
			glob.DownloadFileStoreName, false, "off", false, "off", false,
			false, "off", 0.0, printHeadersData, false,
			false, false, false, Noden)

		time.Sleep(time.Second * 2)

	}
	<-concurrentGoroutines
}
