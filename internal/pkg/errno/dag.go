package errno

import (
	"aimtp/pkg/errorsx"
	"net/http"
)

var (
	ErrDAGNotFound = &errorsx.ErrorX{
		Code:http.StatusNotFound,
		Reason:"NotFound.DAGNotFound",
		Message:"DAG not found.",
	}

	ErrDAGAlreadyExist = &errorsx.ErrorX{
		Code:http.StatusConflict,
		Reason:"Conflict.DAGAlreadyExist",
		Message:"DAG already exist.",
	}

	ErrClusterNotFound = &errorsx.ErrorX{
		Code:http.StatusNotFound,
		Reason:"NotFound.ClusterNotFound",
		Message:"Cluster not found.",
	}

	ErrClusterUnhealthy = &errorsx.ErrorX{
		Code:http.StatusServiceUnavailable,
		Reason:"ServiceUnavailable.ClusterUnhealthy",
		Message:"Cluster is unhealthy.",
	}

	ErrSerializeDAGPayload = &errorsx.ErrorX{
		Code:http.StatusBadRequest,
		Reason:"BadRequest.SerializeDAGPayload",
		Message:"Failed to serialize DAG payload.",
	}

	ErrCreateDAGFailed = &errorsx.ErrorX{
		Code:http.StatusInternalServerError,
		Reason:"InternalServerError.CreateDAGFailed",
		Message:"Failed to create DAG.",
	}
)
