package marid

type settings struct {
	verbose        bool
	bufferPoolSize int
}

func defaultSettings() *settings {
	return &settings{
		verbose:        false,
		bufferPoolSize: 10,
	}
}
