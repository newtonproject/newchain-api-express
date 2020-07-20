package params

const (
	ContentType             = "application/json"
	MaxRequestContentLength = 1024 * 128
)

// wait level
const (
	LevelNoWait        = 0
	LevelWaitBroadcast = 1
	LevelWaitConfirmed = 2
)
