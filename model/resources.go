package model

type (
	BaseResource struct {
		Name      string     `bson:"name" json:"name"`
		Namespace string     `bson:"namespace" json:"namespace"`
		Labels    []KeyValue `bson:"labels" json:"labels"`
	}

	Deployment struct {
		BaseResource
		Replicas      int32      `bson:"replicas" json:"replicas"`
		ImageRegistry string     `bson:"imageRegistry" json:"imageRegistry"`
		CpuLimits     string     `bson:"cpuLimits" json:"cpuLimits"`
		MemoryLimits  string     `bson:"memoryLimits" json:"memoryLimits"`
		Port          int32      `bson:"port" json:"port"`
		Volume        *Volume    `bson:"volumes" json:"volumes"`
		ConfigMap     *ConfigMap `bson:"configMap" json:"configMap"`
	}

	Volume struct {
		Name                  string
		PersistantVolumeClaim *PersistentVolumeClaim `bson:"pvc" json:"pvc"`
		MountPath             string                 `bson:"mountPath" json:"mountPath"`
	}

	PersistentVolumeClaim struct {
		BaseResource
		StorageClassName string `bson:"storageClassName" json:"storageClassName"`
		AccessModes      string `bson:"accessModes" json:"accessModes"`
		StorageSize      int64  `bson:"storageSize" json:"storageSize"`
	}

	ConfigMap struct {
		BaseResource
		Data []KeyValue `bson:"data" json:"data"`
	}

	Service struct {
		BaseResource
		Port         int32         `bson:"port" json:"port"`
		TargetPort   int32         `bson:"targetPort" json:"targetPort"`
		IngressRoute *IngressRoute `bson:"ingressRoute" json:"ingressRoute"`
		Deployment   *Deployment   `bson:"deployment" json:"deployment"`
	}

	IngressRoute struct {
		BaseResource
		Entrypoints []string `bson:"entrypoints" json:"entrypoints"`
		Domain      string   `bson:"domain" json:"domain"`
		Service     *Service `bson:"service" json:"service"`
	}

	ContainerStatus string
)

const (
	ContainerCreatedStatus ContainerStatus = "created"
)

const (
	EnvironmentLabel  = "environment"
	OwnerLabel        = "ownedBy"
	AppLabel          = "app"
	VisibilityLabel   = "visibility"
	ImageLabel        = "image"
	PortLabel         = "port"
	IpaasVersionLabel = "ipaasVersion"
	IpaasManagedLabel = "ipaasManaged"
	ResourceNameLabel = "resourceName"
)
