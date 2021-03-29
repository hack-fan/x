package xlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap/zapcore"
)

// WeworkSender can send notification to wechat work.
type WeworkSender struct {
	BaseURL  string `default:"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="`
	DebugKey string
	WarnKey  string
	ErrorKey string
}

// RobotMsg is message api model
type RobotMsg struct {
	MsgType string  `json:"msgtype"`
	Text    MsgText `json:"text"`
}

// MsgText is text message api model
type MsgText struct {
	Content string `json:"content"`
}

// SendRobotMsg send robot message by wechat work web api
func (s WeworkSender) SendRobotMsg(key, content string) error {
	msg, err := json.Marshal(&RobotMsg{
		MsgType: "text",
		Text:    MsgText{content},
	})
	if err != nil {
		return fmt.Errorf("send wework msg failed: %w", err)
	}
	body := bytes.NewReader(msg)
	req, err := http.NewRequest("POST", s.BaseURL+key, body)
	if err != nil {
		return fmt.Errorf("send wework msg failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpc.Do(req)
	if err != nil {
		return fmt.Errorf("wechat work send robot message api error: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("wechat work send robot message api error: %s", resp.Status)
	}
	return nil
}

func (s WeworkSender) Process(entry zapcore.Entry) error {
	switch entry.Level {
	case zapcore.DebugLevel:
		return s.SendRobotMsg(s.DebugKey, entry.Message)
	case zapcore.WarnLevel:
		return s.SendRobotMsg(s.WarnKey, entry.Message)
	case zapcore.ErrorLevel:
		return s.SendRobotMsg(s.ErrorKey, entry.Message)
	}
	return nil
}
