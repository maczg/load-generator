package conf

import (
	"os"
	"strconv"
)

var (
	ServiceUrl = getEnvString("SERVICE_URL", "http://edge-vp-1.master.particles.dieei.unict.it")
	ZipfS      = getEnvFloat64("ZIPF_S", 1.2)
	ZipfV      = getEnvFloat64("ZIPF_V", 1)
	ExpLambda  = getEnvFloat64("EXP_AVG", 0.1)
)

func getEnvString(key string, defaultValue string) string {
	if str, ok := os.LookupEnv(key); ok {
		return str
	}
	return defaultValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	if str, ok := os.LookupEnv(key); ok {
		if i, e := strconv.ParseFloat(str, 64); e != nil {
			return i
		}
	}
	return defaultValue
}
