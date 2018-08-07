// +build unit

package utils

// SendNonBlocking sends a single value to the given channel without blocking the parent go routine
func SendNonBlocking(value interface{}, channel chan<-interface{}) {
	//channel <- value
	select {
	case channel <- value:
	default:
	}
}