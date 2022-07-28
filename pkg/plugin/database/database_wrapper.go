package database

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/parser"
	"strconv"
	"time"
)

var database *sql.DB = nil

type Credentials struct {
	Hostname string
	User     string
	Password string
	Database string
}

type TestResult struct {
	Success bool
	Message string
}

func IsConnected() bool {
	return database != nil
}

func TestConnection(cred *Credentials) (result *TestResult) {
	log.DefaultLogger.Info("TestConnection called")
	defer func() {
		if r := recover(); r != nil {
			result = new(TestResult)
			result.Success = false
			result.Message = "Database error"
		}
	}()

	// Invalidate current database pool
	if database != nil {
		database.Close()
		database = nil
	}

	cfg := mysql.Config{
		User:                 cred.User,
		Passwd:               cred.Password,
		Net:                  "tcp",
		Addr:                 cred.Hostname,
		DBName:               cred.Database,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return &TestResult{
			Success: false,
			Message: err.Error(),
		}
	}

	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)

	if err != nil {
		db.Close()
		return &TestResult{
			Success: false,
			Message: err.Error(),
		}
	}

	database = db
	return &TestResult{
		Success: true,
		Message: "OK: " + version,
	}
}

func QueryDevices() ([]Device, error) {
	log.DefaultLogger.Info("QueryDevices called")
	if !IsConnected() {
		return nil, errors.New("not connected to any database")
	}
	res, err := database.Query("SELECT id, serial_id, name FROM devices")
	defer res.Close()

	if err != nil {
		log.DefaultLogger.Error("QueryDevices", err)
		return nil, err
	}

	devices := make([]Device, 0)
	for res.Next() {
		var device Device
		err := res.Scan(&device.Id, &device.SerialId, &device.Name)

		if err != nil {
			log.DefaultLogger.Error("QueryDevices", err)
			return nil, err
		}
		devices = append(devices, device)
	}

	log.DefaultLogger.Info("Found "+strconv.Itoa(len(devices))+" devices.", "devices", devices)
	return devices, nil
}

func QueryMetrics(deviceIdsCsv *string) ([]Metric, error) {
	log.DefaultLogger.Info("QueryMetrics called")
	if !IsConnected() {
		return nil, errors.New("not connected to any database")
	}

	query := "SELECT m.id, m.device_id, m.slave_id, m.function_code," +
		" m.register_start, m.data_format, m.byte_order, m.refresh_rate, m.name, d.name" +
		" FROM metrics m JOIN devices d on m.device_id = d.id"
	if deviceIdsCsv != nil {
		query += " WHERE device_id in (" + *deviceIdsCsv + ")"
	}
	res, err := database.Query(query)
	defer res.Close()

	if err != nil {
		log.DefaultLogger.Error("QueryMetrics", err)
		return nil, err
	}

	metrics := make([]Metric, 0)
	for res.Next() {
		var metric Metric
		err := res.Scan(&metric.Id,
			&metric.DeviceId,
			&metric.SlaveId,
			&metric.FunctionCode,
			&metric.RegisterStart,
			&metric.DataFormat,
			&metric.ByteOrder,
			&metric.RefreshRate,
			&metric.Name,
			&metric.DeviceName)

		if err != nil {
			log.DefaultLogger.Error("QueryMetrics", err)
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func QueryMetricsData(metricIdsCsv *string, timerange backend.TimeRange) ([]DeviceWithMetrics, error) {
	log.DefaultLogger.Info("QueryMetricsData called")

	if !IsConnected() {
		return nil, errors.New("not connected to any database")
	}

	// Query metrics
	query := "select d.id, d.name, m.id, m.name, m.data_format, m.byte_order from metrics m" +
		" join devices d on m.device_id = d.id"
	if metricIdsCsv != nil {
		query += " WHERE m.id in (" + *metricIdsCsv + ")"
	}
	res, err := database.Query(query)
	if err != nil {
		log.DefaultLogger.Error("QueryMetricsData", err)
		return nil, err
	}

	metrics := make(map[int64]*MetricWithData)
	devices := make(map[int64]*DeviceWithMetrics)
	for res.Next() {
		var metric Metric
		var device Device
		err := res.Scan(&device.Id, &device.Name, &metric.Id, &metric.Name, &metric.DataFormat, &metric.ByteOrder)
		if err != nil {
			return nil, err
		}

		parser := parser.GetBytesToDoubleParser(metric.DataFormat, metric.ByteOrder)
		metrics[metric.Id] = &MetricWithData{
			Metric: metric,
			Data:   make([]*MetricData, 0),
			parser: parser,
		}

		if _, in := devices[device.Id]; !in {
			devices[device.Id] = &DeviceWithMetrics{
				Device:  device,
				Metrics: make([]*MetricWithData, 0),
			}
		}
		devices[device.Id].Metrics = append(devices[device.Id].Metrics, metrics[metric.Id])
	}

	query = "SELECT id, metric_id, value, UNIX_TIMESTAMP(timestamp) FROM metrics_data"
	query += " WHERE UNIX_TIMESTAMP(timestamp) > " + strconv.FormatInt(timerange.From.Unix(), 10) +
		" AND UNIX_TIMESTAMP(timestamp) < " + strconv.FormatInt(timerange.To.Unix(), 10)
	if metricIdsCsv != nil {
		query += " AND metric_id in (" + *metricIdsCsv + ")"
	}
	query += " ORDER BY timestamp ASC"
	log.DefaultLogger.Info("QUERY " + query)
	res.Close()
	res, err = database.Query(query)
	defer res.Close()

	if err != nil {
		log.DefaultLogger.Error("QueryMetricsData", err)
		return nil, err
	}

	for res.Next() {
		var d MetricData
		var timestamp int64
		err := res.Scan(&d.Id, &d.MetricId, &d.Value, &timestamp)
		d.Timestamp = time.Unix(timestamp, 0)

		if err != nil {
			log.DefaultLogger.Error("QueryMetricsData", err)
			return nil, err
		}
		mwd := metrics[d.MetricId]
		d.NumValue = mwd.parser(d.Value)
		mwd.Data = append(mwd.Data, &d)
	}

	// extract Metrics from map
	data := make([]DeviceWithMetrics, 0, len(devices))
	for _, d := range devices {
		data = append(data, *d)
	}
	return data, nil
}
