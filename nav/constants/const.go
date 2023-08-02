package constants

const (
	OSExitSuccess = 0
	OSExitError   = -1
)

type OutMode int64
type OutIMode int64

const (
	_ OutIMode = iota
	OText
	OPNG
	OJPG
	OSVG
	OutIModeLast
)

// Const values for configuration mode field.
const (
	_ OutMode = iota
	PrintAll
	PrintSubsys
	PrintSubsysWs
	PrintTargeted
	OutModeLast
)

// Const values for output type.
const (
	InvalidOutput int = iota
	GraphOnly
	JsonOutputPlain
	JsonOutputB64
	JsonOutputGZB64
)

// Configuration defaults.
const (
	DefaultMode        = PrintSubsys
	DefaultOutputType  = "graphOnly"
	DefaultGOutputType = 1
	DefaultMaxDepth    = 0
	DefaultDBDriver    = "postgres"
	DefaultDBInstance  = 1
)

// App description.
const (
	AppName  string = "Nav - kernel symbol navigator"
	AppUsage string = "Usage:\n  nav [FLAGS]"
)
