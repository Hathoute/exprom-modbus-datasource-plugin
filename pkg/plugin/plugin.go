package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/database"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/helper"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*SampleDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SampleDatasource)(nil)
	_ backend.StreamHandler         = (*SampleDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SampleDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &SampleDatasource{}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SampleDatasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SampleDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// check connection to the database
	if !database.IsConnected() {
		credentials, err := helper.GetCredentials(req.PluginContext.DataSourceInstanceSettings)
		if err != nil {
			return nil, err
		}
		result := database.TestConnection(credentials)
		if !result.Success {
			return nil, errors.New("cannot connect to database: " + result.Message)
		}
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := &backend.DataResponse{}
		// Unmarshal the JSON into our queryModel.
		var qm queryModel
		res.Error = json.Unmarshal(q.JSON, &qm)
		if res.Error != nil {
			response.Responses[q.RefID] = *res
			continue
		}

		switch qm.Entity {
		case "Devices":
			res = d.handleDevicesQuery(req.PluginContext, q, qm)
			break
		case "Metrics":
			res = d.handleMetricsQuery(req.PluginContext, q, qm)
			break
		case "MetricsData":
			res = d.handleMetricsDataQuery(req.PluginContext, q, qm)
			break
		default:
			res = &backend.DataResponse{
				Error: errors.New("unknown entity '" + qm.Entity + "'"),
			}
		}

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = *res
	}

	return response, nil
}

func (d *SampleDatasource) handleDevicesQuery(pCtx backend.PluginContext, query backend.DataQuery, qm queryModel) *backend.DataResponse {
	response := &backend.DataResponse{}

	devices, err := database.QueryDevices()
	if err != nil {
		response.Error = err
		return response
	}

	// build fields.
	ids := make([]int64, len(devices))
	names := make([]string, len(devices))
	serials := make([]string, len(devices))

	for i, device := range devices {
		ids[i] = device.Id
		names[i] = device.Name
		serials[i] = device.SerialId
	}

	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("name", nil, names),
		data.NewField("serial_id", nil, serials),
	)

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *SampleDatasource) handleMetricsQuery(pCtx backend.PluginContext, query backend.DataQuery, qm queryModel) *backend.DataResponse {
	response := &backend.DataResponse{}

	var deviceIdsCsv *string
	var _ *string // TODO: Use reflection to only return requested fields
	if devices, ok := qm.Parameters["devices"]; ok {
		deviceIdsCsv = &devices
	}
	if fields, ok := qm.Parameters["fields"]; ok {
		_ = &fields
	}

	metrics, err := database.QueryMetrics(deviceIdsCsv)
	if err != nil {
		response.Error = err
		return response
	}

	ids := make([]int64, len(metrics))
	names := make([]string, len(metrics))
	deviceIds := make([]int64, len(metrics))
	slaveIds := make([]int32, len(metrics))
	fcs := make([]int32, len(metrics))
	registers := make([]int32, len(metrics))
	formats := make([]string, len(metrics))
	orders := make([]string, len(metrics))
	rates := make([]int32, len(metrics))

	for i, metric := range metrics {
		ids[i] = metric.Id
		names[i] = metric.DeviceName + " - " + metric.Name
		deviceIds[i] = metric.DeviceId
		slaveIds[i] = metric.SlaveId
		fcs[i] = metric.FunctionCode
		registers[i] = metric.RegisterStart
		formats[i] = metric.DataFormat
		orders[i] = metric.ByteOrder
		rates[i] = metric.RefreshRate
	}

	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("name", nil, names),
		data.NewField("device_id", nil, deviceIds),
		data.NewField("slave_id", nil, slaveIds),
		data.NewField("function_code", nil, fcs),
		data.NewField("register_start", nil, registers),
		data.NewField("data_format", nil, formats),
		data.NewField("byte_order", nil, orders),
		data.NewField("refresh_rate", nil, rates),
	)

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *SampleDatasource) handleMetricsDataQuery(pCtx backend.PluginContext, query backend.DataQuery, qm queryModel) *backend.DataResponse {
	response := &backend.DataResponse{}

	var metricIdsCsv *string
	if metrics, ok := qm.Parameters["metrics"]; ok {
		metricIdsCsv = &metrics
	}

	log.DefaultLogger.Info("METRICDATA time", "f", query.TimeRange.From.String())

	devices, err := database.QueryMetricsData(metricIdsCsv, query.TimeRange)
	if err != nil {
		response.Error = err
		return response
	}

	for _, device := range devices {
		for _, metric := range device.Metrics {
			frame := metricToFrame(&device.Device, metric)

			if qm.WithStreaming {
				channel := live.Channel{
					Scope:     live.ScopeDatasource,
					Namespace: pCtx.DataSourceInstanceSettings.UID,
					Path:      "stream/metric/" + strconv.FormatInt(metric.Metric.Id, 10),
				}
				frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
			}

			response.Frames = append(response.Frames, frame)
		}
	}

	return response
}

func (d *SampleDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response.
	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []int64{10, 20}),
	)

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      "stream",
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (res *backend.CheckHealthResult, _ error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	credentials, err := helper.GetCredentials(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Cannot extract credentials: " + err.Error(),
		}, nil
	}

	result := database.TestConnection(credentials)
	var status = backend.HealthStatusOk
	if !result.Success {
		status = backend.HealthStatusError
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: result.Message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *SampleDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	path := strings.Split(req.Path, "/")
	if len(path) == 3 && path[0] == "stream" {
		switch path[1] {
		case "metric":
			status = backend.SubscribeStreamStatusOK
		default:
			status = backend.SubscribeStreamStatusNotFound
		}
	}

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	path := strings.Split(req.Path, "/")
	metricId := path[2]
	lastFetch := time.Now()

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case <-time.After(5 * time.Second):
			preFetch := time.Now()
			devices, err := database.QueryMetricsData(&metricId, backend.TimeRange{
				From: lastFetch,
				To:   preFetch.Add(time.Minute),
			})

			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

			var frame *data.Frame
			for _, device := range devices {
				for _, metric := range device.Metrics {
					frame = metricToFrame(&device.Device, metric)
				}
			}

			err = sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

			lastFetch = preFetch
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *SampleDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
