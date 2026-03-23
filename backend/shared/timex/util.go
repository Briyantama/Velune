package timex

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func NowUnix() int64 {
	return time.Now().Unix()
}

func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

func NowUnixMicro() int64 {
	return time.Now().UnixMicro()
}