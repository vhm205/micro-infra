// Package utils provides helper functions
package utils

import "log"

func FailOnError(err error, msg string) (err2 error) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
		panic(err)
	}

	return err
}
