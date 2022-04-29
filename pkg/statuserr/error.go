package statuserr

import (
	"fmt"
	"github.com/go-courier/statuserror"
)

func Wrap(statusCode int, err error, msg string) *statuserror.StatusErr {
	if err == nil {
		return nil
	}

	if msg == "" {
		msg = err.Error()
	}

	return statuserror.Wrap(err, statusCode, "-", msg, fmt.Sprintf("%#v", err))
}
