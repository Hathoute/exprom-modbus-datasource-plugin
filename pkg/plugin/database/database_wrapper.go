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

	query := "SELECT id, device_id, slave_id, function_code," +
		" register_start, data_format, byte_order, refresh_rate, name " +
		"FROM metrics"
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
			&metric.Name)

		if err != nil {
			log.DefaultLogger.Error("QueryMetrics", err)
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func QueryMetricsData(metricIdsCsv *string, timerange backend.TimeRange) ([]MetricData, error) {
	log.DefaultLogger.Info("QueryMetricsData called")
	if !IsConnected() {
		return nil, errors.New("not connected to any database")
	}

	// Query metrics
	query := "SELECT id, data_format, byte_order FROM metrics"
	if metricIdsCsv != nil {
		query += " WHERE id in (" + *metricIdsCsv + ")"
	}
	res, err := database.Query(query)
	if err != nil {
		log.DefaultLogger.Error("QueryMetricsData", err)
		return nil, err
	}
	var parsers map[int64]func([]byte) float64
	for res.Next() {
		var metric Metric
		err := res.Scan(&metric.Id, &metric.DataFormat, &metric.ByteOrder)
		if err != nil {
			return nil, err
		}

		parsers[metric.Id] = parser.GetBytesToDoubleParser(metric.DataFormat, metric.ByteOrder)
	}

	query = "SELECT id, metric_id, value, timestamp FROM metrics_data"
	query += " WHERE timestamp > " + strconv.FormatInt(timerange.From.Unix(), 10) +
		" AND timestamp < " + strconv.FormatInt(timerange.To.Unix(), 10)
	if metricIdsCsv != nil {
		query += " AND metric_id in (" + *metricIdsCsv + ")"
	}
	query += " ORDER BY timestamp DESC"
	res.Close()
	res, err = database.Query(query)
	defer res.Close()

	if err != nil {
		log.DefaultLogger.Error("QueryMetricsData", err)
		return nil, err
	}

	data := make([]MetricData, 0)
	for res.Next() {
		var d MetricData
		err := res.Scan(&d.Id, &d.MetricId, &d.Value, &d.Timestamp)

		if err != nil {
			log.DefaultLogger.Error("QueryMetricsData", err)
			return nil, err
		}
		d.NumValue = parsers[d.Id](d.Value)
		data = append(data, d)
	}

	return data, nil
}
