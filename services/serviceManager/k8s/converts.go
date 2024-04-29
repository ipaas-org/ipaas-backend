package k8smanager

import (
	"github.com/ipaas-org/ipaas-backend/model"
)

func convertModelDataToK8sData(data []model.KeyValue) map[string]string {
	k8sData := make(map[string]string)
	for _, d := range data {
		k8sData[d.Key] = d.Value
	}
	return k8sData
}

func convertK8sDataToModelData(data map[string]string) []model.KeyValue {
	var modelData []model.KeyValue
	for k, v := range data {
		modelData = append(modelData, model.KeyValue{
			Key:   k,
			Value: v,
		})
	}
	return modelData
}
