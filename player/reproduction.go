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

func Reproduction(nreq uint64, u uint64, list []models.VideoMetadata, wg *sync.WaitGroup, dryMode bool) {
	defer utils.HandleError()

	log.Printf("[Req#%d] Reproducing video n. %d => %s", nreq, u, list[u].Id)
	if dryMode {
		time.Sleep(time.Second * 2)
	} else {
		url := httpreq.GetVideoUrl(list[u])
		_ = url
		url = "http://staging.massimogollo.it/videofiles/6238f1eca890ed67d7f8cb50/video.mpd"

		mpd, err := httpreq.ReadMPD(url)
		/*		mpd, err := httpreq.ReadMPD(url)*/
		if err != nil {
			return
		}

		//structlist with only one mpd
		structList := []http.MPD{*mpd.Mpd}
		var printHeadersData map[string]string

		var Noden = P2Pconsul.NodeUrl{}
		//TODO codec hardcoded h264
		player.Stream(structList, glob.DebugFile, false, "h264", glob.CodecName, 2160,
			20000, 30, 2, "conventional", url,
			glob.DownloadFileStoreName, false, "off", false, "off", false,
			false, "off", 0.0, printHeadersData, true,
			false, false, false, Noden)
	}

	log.Printf("[Req#%d] End video n. %d => %s", nreq, u, list[u])
	wg.Done()
}
