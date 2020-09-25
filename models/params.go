package models

// Parameters define arguments with which program started
type Parameters struct {
	Sourcefile string // From where to read training text
	Sourcetext string // Training text itself (optional)
	Codelines  bool   // Treat training text as code?
}
