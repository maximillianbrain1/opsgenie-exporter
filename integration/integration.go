package integration

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/integration"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

const (
	namespace = "opsgenie"
	subsystem = "integration"
	labelID   = "id"
	labelName = "name"
	labelType = "type"
)

var (
	suppressedIntegration *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "notifications_suppressed"),
		"Displays whether or not an Integration has its notifications suppressed",
		[]string{
			labelID,
			labelName,
			labelType,
		},
		nil,
	)

	disabledIntegration *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "disabled"),
		"Displays whether or not an Integration is disabled",
		[]string{
			labelID,
			labelName,
			labelType,
		},
		nil,
	)
)

//
type Config struct {
	Config *client.Config
}

//
type Integration struct {
	Client *integration.Client
}

//
func New(config Config) (*Integration, error) {
	if config.Config == nil {
		// TODO: Error.
	}

	client, err := integration.NewClient(config.Config)

	if err != nil {
		return nil, microerror.Mask(err)
	}

	i := &Integration{
		Client: client,
	}

	return i, nil
}

// Collect collects the required metrics via Client
func (i *Integration) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	// channel to send integrations through
	integrations := make(chan integration.GenericFields, 10)

	// Produce integrations onto channel
	g.Go(func() error {
		defer close(integrations)

		listOfIntegrations, err := i.Client.List(ctx)

		if err != nil {
			return microerror.Mask(err)
		}

		for _, i := range listOfIntegrations.Integrations {
			// prevent deadlock if no readers for the channel
			select {
			case <-ctx.Done():
				return ctx.Err()
			case integrations <- i:
			}
		}

		return nil
	})

	// channel to send integrations to find more details on
	integrationIds := make(chan string, 10)

	// Populate disabled metric, and produce ids onto channel
	g.Go(func() error {
		defer close(integrationIds)

		for i := range integrations {

			ch <- prometheus.MustNewConstMetric(
				disabledIntegration,
				prometheus.GaugeValue,
				float64(booleanToInteger(!i.Enabled)),
				i.Id,
				i.Name,
				i.Type,
			)

			// prevent deadlock if no readers for the channel
			select {
			case <-ctx.Done():
				return ctx.Err()
			case integrationIds <- i.Id:
			}
		}

		return nil
	})

	g.Go(func() error {

		for id := range integrationIds {

			ctx := context.Background()
			req := integration.GetRequest{
				Id: id,
			}
			res, err := i.Client.Get(ctx, &req)

			if err != nil {
				return microerror.Mask(err)
			}

			// some integrations do not support suppression or report this field
			if res.Data["suppressNotifications"] != nil {
				ch <- prometheus.MustNewConstMetric(
					suppressedIntegration,
					prometheus.GaugeValue,
					float64(booleanToInteger(res.Data["suppressNotifications"].(bool))),
					id,
					res.Data["name"].(string),
					res.Data["type"].(string),
				)
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

//
func (i *Integration) Describe(ch chan<- *prometheus.Desc) error {
	ch <- disabledIntegration
	ch <- suppressedIntegration

	return nil
}

// booleanToInteger returns 1 or 0 based on true/false argument
func booleanToInteger(b bool) int {
	if b {
		return 1
	}
	return 0
}
