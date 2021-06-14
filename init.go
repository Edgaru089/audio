package audio

// Init initializes OpenAL resources.
//
// It should be called in the main function, preceeding any OpenAL calls.
func Init() {
	initDevice()
}
