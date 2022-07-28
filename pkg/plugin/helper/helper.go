package helper

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/database"
	"unicode"
)

func GetCredentials(instanceSettings *backend.DataSourceInstanceSettings) (*database.Credentials, error) {
	type JSONDataStruct struct {
		Hostname string
		User     string
		Database string
	}
	var jsonData JSONDataStruct

	err := json.Unmarshal(instanceSettings.JSONData, &jsonData)
	if err != nil {
		return nil, err
	}

	// Build Credentials object
	return &database.Credentials{
		Hostname: jsonData.Hostname,
		User:     jsonData.User,
		Password: instanceSettings.DecryptedSecureJSONData["password"],
		Database: jsonData.Database,
	}, nil
}

func SqlFieldToStructField(field string) string {
	structField := ""
	capitalize := true
	for _, c := range field {
		switch c {
		case '_':
			capitalize = true
			break
		case ' ':
			continue
		default:
			if capitalize {
				structField += string(unicode.ToUpper(c))
			} else {
				structField += string(c)
			}
		}
	}
	return structField
}
