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
)
