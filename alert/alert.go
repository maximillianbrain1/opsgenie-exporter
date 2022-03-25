package alert

import (
	"context"
	"github.com/giantswarm/microerror"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

const (
	namespace = "opsgenie"
	subsystem = "alert"

	labelStatus = "status"
)

var (
	alertCount *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "count"),
		"Count of OpsGenie alerts.",
		[]string{
			labelStatus,
		},
		nil,
	)
)

type Config struct {
	Config *client.Config
}

type Alert struct {
	Client *alert.Client
}

func New(config Config) (*Alert, error) {
	if config.Config == nil {
		// TODO: Error.
	}

	alertClient, err := alert.NewClient(config.Config)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	a := &Alert{Client: alertClient}

	return a, nil
}

func (a *Alert) Collect(ch chan<- prometheus.Metric) error {
	var g errgroup.Group

	g.Go(func() error {
		alertRequest := alert.CountAlertsRequest{
			Query: "",
		}

		numAlerts, err := a.Client.CountAlerts(context.Background(), &alertRequest)
		if err != nil {
			return microerror.Mask(err)
		}

		ch <- prometheus.MustNewConstMetric(
			alertCount,
			prometheus.GaugeValue,
			float64(numAlerts.Count),
			"",
		)

		return nil
	})

	g.Go(func() error {
		alertRequest := alert.CountAlertsRequest{
			Query: "status:open",
		}

		numAlerts, err := a.Client.CountAlerts(context.Background(), &alertRequest)
		if err != nil {
			return microerror.Mask(err)
		}
		if err != nil {
			return microerror.Mask(err)
		}
		
		ch <- prometheus.MustNewConstMetric(
			alertCount,
			prometheus.GaugeValue,
			float64(numAlerts.Count),
			"open",
		)

		return nil
	})

	g.Go(func() error {
		alertRequest := alert.CountAlertsRequest{
			Query: "status:closed",
		}

		numAlerts, err := a.Client.CountAlerts(context.Background(), &alertRequest)
		if err != nil {
			return microerror.Mask(err)
		}
		if err != nil {
			return microerror.Mask(err)
		}

		if err != nil {
			return microerror.Mask(err)
		}

		ch <- prometheus.MustNewConstMetric(
			alertCount,
			prometheus.GaugeValue,
			float64(numAlerts.Count),
			"closed",
		)

		return nil
	})

	if err := g.Wait(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Alert) Describe(ch chan<- *prometheus.Desc) error {
	ch <- alertCount

	return nil
}
