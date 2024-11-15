package message

import "encoding/xml"

type ForwardNode struct {
	GroupID    uint32
	SenderID   uint32
	SenderName string
	Time       uint32
	Message    []IMessageElement
}

type (
	MultiMsgLightApp struct {
		App    string `json:"app"`
		Config Config `json:"config"`
		Desc   string `json:"desc"`
		Extra  string `json:"extra"`
		Meta   Meta   `json:"meta"`
		Prompt string `json:"prompt"`
		Ver    string `json:"ver"`
		View   string `json:"view"`
	}

	MultiMsgLightAppExtra struct {
		FileName string `json:"filename"`
		Sum      int    `json:"tsum"`
	}

	Config struct {
		Autosize int64  `json:"autosize"`
		Forward  int64  `json:"forward"`
		Round    int64  `json:"round"`
		Type     string `json:"type"`
		Width    int64  `json:"width"`
	}

	Meta struct {
		Detail Detail `json:"detail"`
	}

	Detail struct {
		News    []News `json:"news"`
		Resid   string `json:"resid"`
		Source  string `json:"source"`
		Summary string `json:"summary"`
		UniSeq  string `json:"uniseq"`
	}

	News struct {
		Text string `json:"text"`
	}
)

type (
	MultiMessage struct {
		XMLName    xml.Name    `xml:"msg"`
		ServiceID  uint        `xml:"serviceID,attr"`
		TemplateID int         `xml:"templateID,attr"`
		Action     string      `xml:"action,attr"`
		Brief      string      `xml:"brief,attr"`
		FileName   string      `xml:"m_fileName,attr"`
		ResID      string      `xml:"m_resid,attr"`
		Total      int         `xml:"tSum,attr"`
		Flag       int         `xml:"flag,attr"`
		Item       MultiItem   `xml:"item"`
		Source     MultiSource `xml:"source"`
	}

	MultiItem struct {
		Layout  int          `xml:"layout,attr"`
		Title   []MultiTitle `xml:"title"`
		Summary MultiSummary `xml:"summary"`
	}

	MultiTitle struct {
		Color string `xml:"color,attr"`
		Size  int    `xml:"size,attr"`
		Text  string `xml:",chardata"`
	}

	MultiSummary struct {
		Color string `xml:"color,attr"`
		Text  string `xml:",chardata"`
	}

	MultiSource struct {
		Name string `xml:"name,attr"`
	}
)
