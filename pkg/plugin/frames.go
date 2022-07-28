package plugin

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/database"
	"time"
)

func metricToFrame(device *database.Device, metric *database.MetricWithData) *data.Frame {
	frame := data.NewFrame(device.Name)

	times := make([]time.Time, len(metric.Data))
	values := make([]float64, len(metric.Data))

	for i, d := range metric.Data {
		times[i] = d.Timestamp
		values[i] = d.NumValue
	}

	// populate fields with metric values
	frame.Fields = append(frame.Fields,
		data.NewField("Value", data.Labels{"metric": metric.Metric.Name}, values),
		data.NewField("Time", nil, times),
	)

	return frame
}
