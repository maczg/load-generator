package player

import (
	log "github.com/sirupsen/logrus"
	httpreq "load-generator/httpreq"
	"load-generator/models"
	"load-generator/utils"
	"sync"
	"time"
)

func reproduction(nreq uint64, u uint64, list []models.VideoMetadata, wg *sync.WaitGroup, dryMode bool) {
	defer utils.HandleError()

	log.Printf("[Req#%d] Reproducing video n. %d => %s", nreq, u, list[u].Id)
	if dryMode {
		time.Sleep(time.Second * 2)
	} else {
		_ = httpreq.GetVideoUrl(list[u])
		//TODO handle error properly
		/*		mpd, meta, _ := dashHttp.ReadURL(videoUrl)
				mpdCodec, mpdCodexIdx, useAudio := dashHttp.GetCodecPerMpd(*mpd)

				player.Stream(structList, glob.DebugFile, debugLog, *codecPtr, glob.CodecName, *maxHeightPtr,
					*streamDurationPtr, *maxBufferPtr, *initBufferPtr, *adaptPtr, *urlPtr,
					fileDownloadLocation, extendPrintLog, *hlsPtr, hlsBool, *quicPtr, quicBool,
					getHeaderBool, *getHeaderPtr, exponentialRatio, printHeadersData, printLog,
					useTestbedBool, getQoEBool, saveFilesBool, Noden)

				waitForEndOfVideo(nreq)*/
	}
	log.Printf("[Req#%d] End video n. %d => %s", nreq, u, list[u])
	wg.Done()
}
