package model

type Template struct {
	Code            string
	Name            string
	ImageName       string
	ImageID         string
	ListeningPort   string
	RequiredEnvs    []KeyValue
	OptionalEnvs    []KeyValue
	IsUpdatabale    bool
	Description     string
	Documentation   string
	PersistancePath string
	Kind            ServiceKind
}

/*
keyvalue{
	key: name of the env var
	value: description of the env var
}
*/
