package imageBuilder

type ImageBuilder struct {
	uri               string
	requestQueueName  string
	responseQueueName string
}

func NewImageBuilder(uri, requestQueue, reponseQueue string) *ImageBuilder {
	return &ImageBuilder{
		uri:               uri,
		requestQueueName:  requestQueue,
		responseQueueName: reponseQueue,
	}
}

func (i ImageBuilder) BuildImage(uuid, providerToken, userID, repo, branch string) error
