package model

type Template struct {
	Code            string      `bson:"code" json:"code"`
	Available       bool        `bson:"available" json:"available"`
	Name            string      `bson:"name" json:"name"`
	ImageName       string      `bson:"imageName" json:"-"`
	ImageID         string      `bson:"imageID" json:"-"`
	ListeningPort   string      `bson:"listeningPort" json:"-"`
	RequiredEnvs    []KeyValue  `bson:"requiredEnvs" json:"requiredEnvs"`
	OptionalEnvs    []KeyValue  `bson:"optionalEnvs" json:"optionalEnvs"`
	DefaultEnvs     []KeyValue  `bson:"defaultEnvs" json:"-"`
	IsUpdatabale    bool        `bson:"isUpdatabale" json:"-"`
	Description     string      `bson:"description" json:"description"`
	Documentation   string      `bson:"documentation" json:"documentation"`
	PersistancePath string      `bson:"persistancePath" json:"-"`
	Kind            ServiceKind `bson:"kind" json:"kind"`
}

/*
keyvalue{
	key: name of the env var
	value: description of the env var
}
*/
