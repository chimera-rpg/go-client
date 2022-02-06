package audio

// CommandI is the interface for our audio Command messages.
type CommandI interface {
}

// CommandNewSound adds a sound from bytes matching the given ID.
type CommandNewSound struct {
	ID       uint32
	Type     uint8 // See network sound command type
	Filepath string
}

// CommandPlaySound starts playing sounds matching the given ID.
type CommandPlaySound struct {
	ID             uint32
	Volume         float32    // 0-1
	ChannelVolumes [8]float64 // I have no idea about this.
}

// CommandStopSound stops playing all sounds matching ID.
type CommandStopSound struct {
	ID uint32
}

// CommandPlayMusic starts playing music matching the given ID.
type CommandPlayMusic struct {
	ID             uint32
	PlaybackID     uint32
	Volume         float32    // 0-1
	ChannelVolumes [8]float64 // I have no idea about this.
}

// CommandStopMusic stops playing music matching the given ID.
type CommandStopMusic struct {
	PlaybackID uint32
}
