package main

import (
	"flag"
	"log"
	"net/http"

	i2c "github.com/d2r2/go-i2c"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// https://github.com/PiSupply/PiJuice/blob/65f7a66c1a3e24eed87c16a98c26ec19dd41e871/Software/Source/pijuice.py#L140-L153
const (
	CHARGE_LEVEL_CMD        = 0x41
	BATTERY_TEMPERATURE_CMD = 0x47
	// BATTERY_VOLTAGE_CMD     = 0x49
	// IO_VOLTAGE_CMD          = 0x4d
)

const namespace = "pijuice"

var (
	chargeLevel = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "charge_level"),
		"Charge level",
		nil,
		nil,
	)
	batteryTemperature = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "battery_temperature"),
		"Battery temperature (Celsius)",
		nil,
		nil,
	)
	batteryVoltage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "battery_voltage"),
		"Battery voltage",
		nil,
		nil,
	)
	ioVoltage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "io_voltage"),
		"IO voltage",
		nil,
		nil,
	)
)

type Exporter struct {
	i2cAddress uint8
	i2cBus     int
}

func NewExporter(i2cAddress uint8, i2cBus int) *Exporter {
	return &Exporter{
		i2cAddress,
		i2cBus,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- chargeLevel
	ch <- batteryTemperature
	ch <- batteryVoltage
	ch <- ioVoltage
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	i2c, err := i2c.NewI2C(e.i2cAddress, e.i2cBus)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	// TODO: Handle errors, add more metrics
	chargeLvl, _, _ := i2c.ReadRegBytes(CHARGE_LEVEL_CMD, 1)
	ch <- prometheus.MustNewConstMetric(chargeLevel, prometheus.GaugeValue, float64(chargeLvl[0]))

	batteryTemp, _, _ := i2c.ReadRegBytes(BATTERY_TEMPERATURE_CMD, 2)
	ch <- prometheus.MustNewConstMetric(batteryTemperature, prometheus.GaugeValue, float64(batteryTemp[0]))
}

func main() {
	var bind string

	flag.StringVar(&bind, "bind", ":9886", "bind")
	flag.Parse()

	exporter := NewExporter(0x14, 1)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
             <head><title>PiJuice Exporter</title></head>
             <body>
             <h1>PiJuice Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})

	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatal(err)
	}
}
