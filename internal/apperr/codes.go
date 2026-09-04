package apperr

const (
	CodeInternal = "INTERNAL"

	CodeMediaNoAudioStream = "MEDIA_NO_AUDIO_STREAM"
	CodeMediaCorrupt       = "MEDIA_CORRUPT"
	CodeMediaDRMProtected  = "MEDIA_DRM_PROTECTED"
	CodeMediaUnsupported   = "MEDIA_UNSUPPORTED"

	CodeDiskInsufficientSpace = "DISK_INSUFFICIENT_SPACE"

	CodeModelNotInstalled     = "MODEL_NOT_INSTALLED"
	CodeModelChecksumMismatch = "MODEL_CHECKSUM_MISMATCH"
	CodeModelInUse            = "MODEL_IN_USE"

	CodeTranscribeOutOfMemory = "TRANSCRIBE_OUT_OF_MEMORY"
	CodeTranscribeFailed      = "TRANSCRIBE_FAILED"

	CodeNetworkUnavailable = "NETWORK_UNAVAILABLE"
)

// Read by someone who does not know what a codec is. Keep them that way.
var messages = map[string]string{
	CodeInternal:              "Something went wrong. The details are in the log file.",
	CodeMediaNoAudioStream:    "This file has no audio to transcribe.",
	CodeMediaCorrupt:          "This file appears to be damaged and cannot be read.",
	CodeMediaDRMProtected:     "This file is copy protected and cannot be opened.",
	CodeMediaUnsupported:      "This file format is not supported.",
	CodeDiskInsufficientSpace: "Not enough disk space to process this file.",
	CodeModelNotInstalled:     "That model has not been downloaded yet.",
	CodeModelChecksumMismatch: "The download was damaged. Try again.",
	CodeModelInUse:            "This model is in use and cannot be removed.",
	CodeTranscribeOutOfMemory: "Not enough memory for this model. Try a smaller one.",
	CodeTranscribeFailed:      "Transcription failed. The details are in the log file.",
	CodeNetworkUnavailable:    "Could not reach the download server. Check your connection.",
}

func Message(code string) string {
	if m, ok := messages[code]; ok {
		return m
	}
	return messages[CodeInternal]
}
