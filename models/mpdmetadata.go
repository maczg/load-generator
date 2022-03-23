package models

import (
	"net/http/httptrace"
	"time"
)

type DNSInfo struct {
	DnsStartInfo    httptrace.DNSStartInfo
	DnsStartTime    time.Time
	DnsDoneInfo     httptrace.DNSDoneInfo
	DnsDoneDuration time.Duration
}

type ConnectionInfo struct {
	StartNetwork, StartAddr string
	EndNetwork, EndAddr     string
	ConnStartTime           time.Time
	ConnEndTime             time.Time
	Duration                time.Duration
	Err                     error
}

type MpdMetadata struct {
	HostPort      string
	DNSInfo       DNSInfo
	ConnInfo      ConnectionInfo
	GotConnection httptrace.GotConnInfo
	RTT2FirstByte time.Duration
}

func NewMpdMetadata() *MpdMetadata {
	return &MpdMetadata{
		HostPort: "",
		DNSInfo: DNSInfo{
			DnsStartInfo:    httptrace.DNSStartInfo{},
			DnsStartTime:    time.Time{},
			DnsDoneInfo:     httptrace.DNSDoneInfo{},
			DnsDoneDuration: 0,
		},
		ConnInfo:      ConnectionInfo{},
		GotConnection: httptrace.GotConnInfo{},
		RTT2FirstByte: 0,
	}
}

//to test GETMpd
//mpd, err := httpreq.ReadMPD("http://staging.massimogollo.it/videofiles/62365d47a890ed67d7f8cb49/video.mpd")

//TODO check others usefull callback
func GetMpdHttpTrace(mpdMeta *MpdMetadata, startTime *time.Time) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			mpdMeta.HostPort = hostPort
		},
		GotFirstResponseByte: func() {
			mpdMeta.RTT2FirstByte = time.Since(*startTime)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			mpdMeta.GotConnection = info
		},
		Got100Continue: nil,
		Got1xxResponse: nil,
		DNSStart: func(info httptrace.DNSStartInfo) {
			mpdMeta.DNSInfo.DnsStartInfo = info
			mpdMeta.DNSInfo.DnsStartTime = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			mpdMeta.DNSInfo.DnsDoneInfo = info
			mpdMeta.DNSInfo.DnsDoneDuration = time.Since(mpdMeta.DNSInfo.DnsStartTime)
		},
		ConnectStart: func(network, addr string) {
			mpdMeta.ConnInfo.ConnStartTime = time.Now()
			mpdMeta.ConnInfo.StartAddr = addr
			mpdMeta.ConnInfo.StartNetwork = network
		},
		ConnectDone: func(network, addr string, err error) {
			mpdMeta.ConnInfo.EndNetwork = network
			mpdMeta.ConnInfo.EndAddr = addr
			mpdMeta.ConnInfo.ConnEndTime = time.Now()
			mpdMeta.ConnInfo.Duration = time.Since(mpdMeta.ConnInfo.ConnStartTime)
		},
		TLSHandshakeStart: nil,
		TLSHandshakeDone:  nil,
		WroteHeaderField:  nil,
		WroteHeaders:      nil,
		Wait100Continue:   nil,
		WroteRequest:      nil,
	}

}
