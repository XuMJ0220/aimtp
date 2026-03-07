package errno

import (
	"aimtp/pkg/errorsx"
	"net/http"
)

var (
	ErrPodNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.PodNotFound",
		Message: "Pod not found.",
	}
)
