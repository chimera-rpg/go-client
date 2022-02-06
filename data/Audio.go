package data

// Audio provides an AudioID and SoundSetIDS->Sounds pairing.
type Audio struct {
	AudioID   uint32
	SoundSets map[uint32][]Sound
	Pending   bool
}

// Sound provdes a SoundID and Text pairing.
type Sound struct {
	SoundID uint32
	Text    string
	Pending bool
}

// SoundEntry provides something.
type SoundEntry struct {
	//Bytes   []byte // The underlying byte data for the sound.
	Filepath string // Path to the file on distk.
	Type     uint8  // See network sound command type
	Pending  bool   // If the sound has been received yet.
}
