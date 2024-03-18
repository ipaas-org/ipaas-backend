package k8smanager

import (
	"github.com/ipaas-org/ipaas-backend/model"
)

// func convertModelEnvToK8sEnv(envs []model.KeyValue) []v1.EnvVar {
// 	var k8sEnvs []v1.EnvVar
// 	for _, env := range envs {
// 		k8sEnvs = append(k8sEnvs, v1.EnvVar{
// 			Name:  env.Key,
// 			Value: env.Value,
// 		})
// 	}
// 	return k8sEnvs
// }

func convertModelKeyValuesToLables(labels []model.KeyValue) map[string]string {
	k8sLabels := make(map[string]string)
	for _, l := range labels {
		k8sLabels[l.Key] = l.Value
	}
	return k8sLabels
}

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
