package structer

const (
	LogTypeSet = "typeset"

	// Log code indicating an error occurred in the types.Config.Error
	// callback, but execution continued.
	LogTypesConfigError = 1

	// Log code indicating an error was returned by types.Config.Check, but
	// execution continued.
	LogTypeCheck = 2
)

type Log interface {
	Log(category string, code int, message string, args ...interface{})
}

func wlog(log Log, category string, code int, message string, args ...interface{}) {
	if log != nil {
		log.Log(category, code, message, args...)
	}
}
