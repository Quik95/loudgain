package loudgain

type WriteMode int

const (
	DeleteTags WriteMode = iota
	WriteRG2
	ExtraTags
	ExtraTagsLU
	SkipWritingTags
	InvalidWriteMode
)

func StringToWriteMode(in string) WriteMode {
	switch in {
	case "d":
		return DeleteTags
	case "i":
		return WriteRG2
	case "e":
		return ExtraTags
	case "l":
		return ExtraTagsLU
	case "s":
		return SkipWritingTags
	default:
		return InvalidWriteMode
	}
}
