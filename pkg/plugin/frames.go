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
		values[i] = d.Value
	}

	valueField := data.NewField("Value", data.Labels{"metric": metric.Metric.Name}, values)
	timeField := data.NewField("Time", nil, times)

	// Add field config (unit, ...)
	valueField.Config = &data.FieldConfig{
		Unit: metric.Metric.Unit,
	}

	// populate fields with metric values
	frame.Fields = append(frame.Fields, valueField, timeField)

	return frame
}
