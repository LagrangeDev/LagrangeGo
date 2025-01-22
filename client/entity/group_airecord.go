package entity

type ChatType uint32

const (
	ChatTypeVoice ChatType = 1
	ChatTypeSong  ChatType = 2
)

type (
	AiCharacter struct {
		Name     string `json:"name"`
		VoiceId  string `json:"voice_id"`
		VoiceUrl string `json:"voice_url"`
	}

	AiCharacterInfo struct {
		Type       string        `json:"type"`
		Characters []AiCharacter `json:"characters"`
	}

	AiCharacterList struct {
		Type ChatType
		List []AiCharacterInfo
	}
)
