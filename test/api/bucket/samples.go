package bucket

import "time"

// Struct to hold bucket sample data.
type BucketSample struct {
	Path string
}

type BucketReaperSample struct {
	Path       string
	Expiration *time.Time
}

var Samples = []BucketSample{
	{
		Path: "sample/template_application_import.csv",
	},
}

var ReaperSamples = []BucketReaperSample{
	{
		Path: "reaper",
		Expiration: func() *time.Time {
			expiration := time.Now()
			return &expiration
		}(),
	},
}
