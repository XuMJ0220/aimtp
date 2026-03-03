package event

type DAGCreatedEvent struct {
	DagID     int64   `json:"dag_id"`
	DagName   string  `json:"dag_name"`
	Cluster   string  `json:"cluster"`
	UserName  string  `json:"user_name"`
	QueueName *string `json:"queue_name"`
	Engine    *string `json:"engine"`
}
