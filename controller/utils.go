package controller

import "github.com/ipaas-org/ipaas-backend/model"

func convertModelKeyValueToMap(model []model.KeyValue) map[string]string {
	m := make(map[string]string)
	for _, kv := range model {
		m[kv.Key] = kv.Value
	}
	return m
}
