package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/v52/common"
	"github.com/oracle/oci-go-sdk/v52/monitoring"
)

type monitoringController struct {
	client    *monitoring.MonitoringClient
	initiated bool
}

func newMonitoringController() *monitoringController {
	return &monitoringController{
		client:    nil,
		initiated: false,
	}
}

func (controller *monitoringController) init(configProvider *common.ConfigurationProvider) error {
	if c, err := monitoring.NewMonitoringClientWithConfigurationProvider(*configProvider); err == nil {
		controller.client = &c
		controller.initiated = true
		return nil
	} else {
		controller.initiated = false
		return err
	}
}

func JoinMetricsQueryString(metric string, interval string, instanceId string, groupingFunction string) string {
	return fmt.Sprintf("%s[%s]{resourceId=%s}.%s()", metric, interval, instanceId, groupingFunction)
}

// metric[interval]{resourceId="resourceId"}.groupingfunction.statistic
// metric : CpuUtilization, MemoryUtilization
// interval : 1-59m 1-23h
// instaceId : instance ocid
// groupingFunction : sum, avg, max, min
// compartmentId : compartment ocid
// start date
// stop date
// returns map[timestamp] = value both float64

func (controller *monitoringController) getMetrics(
	ctx context.Context,
	metric string,
	interval string,
	instanceId string,
	groupingFunction string,
	compartmentId string,
	startDate time.Time,
	stopDate time.Time) (map[float64]float64, error) {
	query := JoinMetricsQueryString(metric, interval, instanceId, groupingFunction)

	return controller.getMetricsByQuery(ctx, query, compartmentId, startDate, stopDate)

}

func (controller *monitoringController) getMetricsByQuery(
	ctx context.Context,
	query string,
	compartmentId string,
	startDate time.Time,
	stopDate time.Time) (map[float64]float64, error) {
	req := monitoring.SummarizeMetricsDataRequest{
		CompartmentId: common.String(compartmentId),
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			EndTime:   &common.SDKTime{Time: stopDate},
			Namespace: common.String("oci_computeagent"),
			Query:     common.String(query),
			StartTime: &common.SDKTime{Time: startDate}}}

	// Send the request using the service client
	resp, err := controller.client.SummarizeMetricsData(context.Background(), req)

	if err != nil {
		return nil, err
	}

	if len(resp.Items) > 1 {
		return nil, fmt.Errorf("number of Items returned should be 1, got %d", len(resp.Items))
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	result := make(map[float64]float64)

	for _, item := range resp.Items[0].AggregatedDatapoints {
		t := float64(item.Timestamp.Time.Unix())
		v := item.Value
		result[t] = *v
	}
	return result, nil
}

// TODO
