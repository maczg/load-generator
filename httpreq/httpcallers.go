package httpreq

import (
	"fmt"
	"io/ioutil"
	"load-generator/conf"
	"load-generator/models"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

func GetVideoUrl(v models.VideoMetadata) string {
	mpdUrl := fmt.Sprintf("%s/vms/videos/%s", conf.ServiceUrl, v.Id)
	return mpdUrl
}

func ReadMPD(url string) (mpd *models.MpdFull, err error) {
	mpd = models.NewMpdFull()
	var startTime time.Time

	var request *http.Request

	url = strings.TrimSpace(url)

	if request, err = http.NewRequest("GET", url, nil); err != nil {
		return nil, err
	}

	//TRACE MPD METRICS
	clientTrace := models.GetMpdHttpTrace(mpd.MpdMetadata, &startTime)
	clientTraceCtx := httptrace.WithClientTrace(request.Context(), clientTrace)
	request = request.WithContext(clientTraceCtx)
	startTime = time.Now()

	client := NewHttpClient()
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	//parse resp in mpd
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	mpd.Mpd, err = models.ParseMpd(body)
	if err != nil {
		return nil, err
	}
	return mpd, nil
}
