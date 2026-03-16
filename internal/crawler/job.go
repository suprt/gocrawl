package crawler

type Job struct {
	URL     string
	Index   int
	Retries int
}

func NewJob(url string, index int) Job {
	return Job{
		URL:     url,
		Index:   index,
		Retries: 0,
	}
}

func (j Job) WithRetry() Job {
	j.Retries++
	return j
}
