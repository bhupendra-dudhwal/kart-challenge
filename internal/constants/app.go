package constants

type Environment string

const (
	ENV_LOCAL       Environment = "local"
	ENV_DEVELOPMENT Environment = "development"
	ENV_PRODUCTION  Environment = "production"
)

func (key Environment) String() string {
	return string(key)
}

func (key Environment) IsValid() bool {
	switch key {
	case ENV_LOCAL, ENV_DEVELOPMENT, ENV_PRODUCTION:
		return true
	default:
		return false
	}
}

type LogLevel string

const (
	LINFO  LogLevel = "info"
	LERROR LogLevel = "error"
	LWARN  LogLevel = "warn"
	LDEBUG LogLevel = "debug"
)

func (key LogLevel) String() string {
	return string(key)
}

func (key LogLevel) IsValid() bool {
	switch key {
	case LINFO, LERROR, LWARN, LDEBUG:
		return true
	default:
		return false
	}
}

type ContentType string

const (
	JSON ContentType = "application/json"
)

func (key ContentType) String() string {
	return string(key)
}

type Compression string

const (
	GZIP Compression = "gzip"
)

func (key Compression) String() string {
	return string(key)
}

type HeaderKey string

const (
	API_KEY HeaderKey = "api_key"
)

func (key HeaderKey) String() string {
	return string(key)
}

type AllowedCharacters string

const (
	Uppercase    AllowedCharacters = "uppercase"
	Alphanumeric AllowedCharacters = "alphanumeric"
	Lowercase    AllowedCharacters = "lowercase"
	Letters      AllowedCharacters = "letters"
	Digits       AllowedCharacters = "digits"
)

func (key AllowedCharacters) IsValid() bool {
	switch key {
	case Uppercase, Alphanumeric, Letters, Lowercase, Digits:
		return true
	default:
		return false
	}
}

type ProcesssStep string

const (
	ADD    ProcesssStep = "add"
	UPDATE ProcesssStep = "update"
)
