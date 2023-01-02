package server

import "time"

func ParseTimeOrNil(timeString string) (*time.Time, error) {
	beganAt, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return nil, err
	}
	return &beganAt, nil
}
