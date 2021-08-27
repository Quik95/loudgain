package loudgain

var (
	ReferenceLoudness LoudnessUnit
	TrackPeakLimit    Decibel
	Pregain           LoudnessUnit
	TagMode           WriteMode
	NoClip            bool
	FFmpegPath        string
)
