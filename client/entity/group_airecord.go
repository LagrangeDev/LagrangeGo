package entity

type ChatType uint32

const (
	ChatTypeVoice ChatType = 1
	ChatTypeSong  ChatType = 2
)

type (
	AiCharacter struct {
		Name     string `json:"name"`
		VoiceID  string `json:"voice_id"`
		VoiceURL string `json:"voice_url"`
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
