package mp3lib

func NewMP3Library(baseDir string) *MP3Library {
	return &MP3Library{
		baseDir: baseDir,
	}
}
