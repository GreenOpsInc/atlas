package util

import (
	"errors"
	"net/http"
)

func CheckResponseStatus(res *http.Response) error {
	switch res.StatusCode {
	case http.StatusBadRequest:
		return errors.New("returned with bad request")
	case http.StatusNotFound:
		return errors.New("returned with not found")
	case http.StatusInternalServerError:
		return errors.New("returned with internal server error")
	case http.StatusServiceUnavailable:
		return errors.New("service is down")
	}
	return nil
}
