package error_wrapper

import (
	"errors"
	"fmt"
)

func Error(funcName string, err error) error {
	return fmt.Errorf("%s->%w", funcName, err)
}

func GetOriginError(err error) error {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	return err
}
