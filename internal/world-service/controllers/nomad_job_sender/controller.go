package nomad_job_sender

import "github.com/gin-gonic/gin"

type NomadJobSenderController interface {
	// StartNewJob starts a job for a specific world and zone.
	StartNewJob(c *gin.Context)
}
