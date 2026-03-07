package errno

import (
	"aimtp/pkg/errorsx"
	"net/http"
)

var (
	ErrJobNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.JobNotFound",
		Message: "Job not found.",
	}
)
