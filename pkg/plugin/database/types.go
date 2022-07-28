package database

import "time"

type Device struct {
	Id       int64
	Name     string
	SerialId string
}

type Metric struct {
	Id            int64
	DeviceId      int64
	DeviceName    string
	Name          string
	SlaveId       int32
	FunctionCode  int32
	RegisterStart int32
	DataFormat    string
	ByteOrder     string
	RefreshRate   int32
}

type MetricData struct {
	Id        int64
	MetricId  int64
	Value     []byte
	Timestamp time.Time

	NumValue float64
}

type MetricWithData struct {
	Metric Metric
	Data   []*MetricData

	parser func([]byte) float64
}

type DeviceWithMetrics struct {
	Device  Device
	Metrics []*MetricWithData
}