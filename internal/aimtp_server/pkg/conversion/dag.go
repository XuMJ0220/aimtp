package conversion

import (
	"aimtp/internal/aimtp_server/model"
)

type DagCreatedEvent struct {
	DagID     int64   `json:"dag_id"`
	DagName   string  `json:"dag_name"`
	Cluster   string  `json:"cluster"`
	UserName  string  `json:"user_name"`
	QueueName *string `json:"queue_name"`
	Engine    *string `json:"engine"`
}

func DagStatusSummaryModelToPublicDagCreatedEvent(model *model.DagStatusSummaryM) *DagCreatedEvent {
	return &DagCreatedEvent{
		DagID:     model.DagID,
		DagName:   model.DagName,
		Cluster:   model.Cluster,
		UserName:  model.UserName,
		QueueName: model.QueueName,
		Engine:    model.Engine,
	}
}
