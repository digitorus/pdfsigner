package webapi

// hanldeScheduleResponse represents response for handleSignSchedule.
type hanldeScheduleResponse struct {
	JobID string `json:"job_id"`
}

// jobStatusResponse represents response of the job status.
type jobStatusResponse struct {
	Job   job    `json:"job"`
	Tasks []task `json:"tasks"`
}

// job is a part of jobStatusResponse.
type job struct {
	ID string `json:"id"`
}

// task is a part of jobStatusResponse.
type task struct {
	ID               string `json:"id"`
	OriginalFileName string `json:"file_name"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
}
