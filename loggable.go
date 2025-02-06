package serr

// Loggable indicates that the implementing type's instances build their own log representation.
type Loggable interface {
	LogString() string
}
