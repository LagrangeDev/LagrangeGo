package message

import (
	"encoding/json"
	"fmt"
)

type elementJSON struct {
	Type ElementType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

func parseJSONMessageElement(elemJSON elementJSON) (IMessageElement, error) {
	var elem IMessageElement
	switch elemJSON.Type {
	case Text:
		var textElem *TextElement
		if err := json.Unmarshal(elemJSON.Data, &textElem); err != nil {
			return elem, fmt.Errorf("解析TextElement失败: %w", err)
		}
		elem = textElem
	case Image:
		var imgElem *ImageElement
		if err := json.Unmarshal(elemJSON.Data, &imgElem); err != nil {
			return elem, fmt.Errorf("解析ImageElement失败: %w", err)
		}
		elem = imgElem
	case At:
		var atElem *AtElement
		if err := json.Unmarshal(elemJSON.Data, &atElem); err != nil {
			return elem, fmt.Errorf("解析AtElement失败: %w", err)
		}
		elem = atElem
	case Reply:
		var replyElem *ReplyElement
		if err := json.Unmarshal(elemJSON.Data, &replyElem); err != nil {
			return elem, fmt.Errorf("解析ReplyElement失败: %w", err)
		}
		elem = replyElem
	case Service:
		var serviceElem *XMLElement
		if err := json.Unmarshal(elemJSON.Data, &serviceElem); err != nil {
			return elem, fmt.Errorf("解析XMLElement失败: %w", err)
		}
		elem = serviceElem
	case Forward:
		var forwardElem *ForwardMessage
		if err := json.Unmarshal(elemJSON.Data, &forwardElem); err != nil {
			return elem, fmt.Errorf("解析ForwardMessage失败: %w", err)
		}
		elem = forwardElem
	case File:
		var fileElem *FileElement
		if err := json.Unmarshal(elemJSON.Data, &fileElem); err != nil {
			return elem, fmt.Errorf("解析FileElement失败: %w", err)
		}
		elem = fileElem
	case Voice:
		var voiceElem *VoiceElement
		if err := json.Unmarshal(elemJSON.Data, &voiceElem); err != nil {
			return elem, fmt.Errorf("解析FileElement失败: %w", err)
		}
		elem = voiceElem
	case Video:
		var videoElem *ShortVideoElement
		if err := json.Unmarshal(elemJSON.Data, &videoElem); err != nil {
			return elem, fmt.Errorf("解析FileElement失败: %w", err)
		}
		elem = videoElem
	case LightApp:
		var lightAppElem *LightAppElement
		if err := json.Unmarshal(elemJSON.Data, &lightAppElem); err != nil {
			return elem, fmt.Errorf("解析FileElement失败: %w", err)
		}
		elem = lightAppElem
	case RedBag:
		return elem, fmt.Errorf("未实现的元素类型: %v", elemJSON.Type)
	case MarketFace:
		var marketFaceElem *MarketFaceElement
		if err := json.Unmarshal(elemJSON.Data, &marketFaceElem); err != nil {
			return elem, fmt.Errorf("解析FileElement失败: %w", err)
		}
		elem = marketFaceElem
	default:
		return elem, fmt.Errorf("未知的元素类型: %v", elemJSON.Type)
	}
	return elem, nil
}

func (g GroupMessage) MarshalJSON() ([]byte, error) {
	type Temp GroupMessage
	temp := struct {
		*Temp
		Elements []elementJSON
	}{
		Temp:     (*Temp)(&g),
		Elements: make([]elementJSON, 0, len(g.Elements)),
	}

	for _, elem := range g.Elements {
		if elem == nil {
			continue
		}
		elemData, err := json.Marshal(elem)
		if err != nil {
			return nil, fmt.Errorf("序列化元素失败: %w", err)
		}
		temp.Elements = append(temp.Elements, elementJSON{
			Type: elem.Type(),
			Data: elemData,
		})
	}

	return json.Marshal(temp)
}

func (g *GroupMessage) UnmarshalJSON(data []byte) error {
	type Temp GroupMessage
	var temp struct {
		*Temp
		Elements []elementJSON
	}
	temp.Temp = (*Temp)(g)

	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("解析GroupMessage失败: %w", err)
	}

	g.Elements = make([]IMessageElement, 0, len(temp.Elements))
	for _, elemJSON := range temp.Elements {
		elem, err := parseJSONMessageElement(elemJSON)
		if err != nil {
			return err
		}
		g.Elements = append(g.Elements, elem)
	}

	g.ID = temp.ID
	g.InternalID = temp.InternalID
	g.GroupUin = temp.GroupUin
	g.GroupName = temp.GroupName
	g.Sender = temp.Sender
	g.Time = temp.Time
	g.OriginalObject = temp.OriginalObject

	return nil
}

func (r ReplyElement) MarshalJSON() ([]byte, error) {
	type Temp ReplyElement
	temp := struct {
		*Temp
		Elements []elementJSON
	}{
		Temp:     (*Temp)(&r),
		Elements: make([]elementJSON, 0, len(r.Elements)),
	}

	for _, elem := range r.Elements {
		if elem == nil {
			continue
		}
		elemData, err := json.Marshal(elem)
		if err != nil {
			return nil, fmt.Errorf("序列化元素失败: %w", err)
		}
		temp.Elements = append(temp.Elements, elementJSON{
			Type: elem.Type(),
			Data: elemData,
		})
	}

	return json.Marshal(temp)
}

func (r *ReplyElement) UnmarshalJSON(data []byte) error {
	type Temp ReplyElement
	var temp struct {
		*Temp
		Elements []elementJSON
	}
	temp.Temp = (*Temp)(r)

	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("解析GroupMessage失败: %w", err)
	}

	r.Elements = make([]IMessageElement, 0, len(temp.Elements))
	for _, elemJSON := range temp.Elements {
		elem, err := parseJSONMessageElement(elemJSON)
		if err != nil {
			return err
		}
		r.Elements = append(r.Elements, elem)
	}

	r.ReplySeq = temp.ReplySeq
	r.SenderUin = temp.SenderUin
	r.SenderUID = temp.SenderUID
	r.GroupUin = temp.GroupUin
	r.Time = temp.Time

	return nil
}

func (f ForwardNode) MarshalJSON() ([]byte, error) {
	type Temp ForwardNode
	temp := struct {
		*Temp
		Elements []elementJSON
	}{
		Temp:     (*Temp)(&f),
		Elements: make([]elementJSON, 0, len(f.Message)),
	}

	for _, elem := range f.Message {
		if elem == nil {
			continue
		}
		elemData, err := json.Marshal(elem)
		if err != nil {
			return nil, fmt.Errorf("序列化元素失败: %w", err)
		}
		temp.Elements = append(temp.Elements, elementJSON{
			Type: elem.Type(),
			Data: elemData,
		})
	}

	return json.Marshal(temp)
}

func (f *ForwardNode) UnmarshalJSON(data []byte) error {
	type Temp ForwardNode
	var temp struct {
		*Temp
		Elements []elementJSON
	}
	temp.Temp = (*Temp)(f)

	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("解析GroupMessage失败: %w", err)
	}

	f.Message = make([]IMessageElement, 0, len(temp.Elements))
	for _, elemJSON := range temp.Elements {
		elem, err := parseJSONMessageElement(elemJSON)
		if err != nil {
			return err
		}
		f.Message = append(f.Message, elem)
	}

	f.GroupID = temp.GroupID
	f.SenderID = temp.SenderID
	f.SenderName = temp.SenderName
	f.Time = temp.Time

	return nil
}
