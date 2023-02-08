package error_handler

import (
	"errors"
)

func ErrorHandler(err error, errs ...[]error) bool {
	for i := 0; i < len(errs); i++ {
		for j := 0; j < len(errs[i]); j++ {
			if errors.Is(err, errs[i][j]) {
				return true
			}
		}
	}
	return false
}
