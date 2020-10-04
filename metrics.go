package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func StartMetricsServer(r *prometheus.Registry) {
	go func() {
		for {
			time.Sleep(time.Second * 5)
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			heapUsage.With(prometheus.Labels{"container": os.Getenv(ContainerNickname)}).Set(float64(mem.Alloc))
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
		log.Fatal(http.ListenAndServe(":3031", nil))
	}()
}
