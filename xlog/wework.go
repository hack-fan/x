package xlog

import (
	"bytes"
	"encoding/json"
	"fmt"

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
	if key == "" {
		return nil
	}
	msg, err := json.Marshal(&RobotMsg{
		MsgType: "text",
		Text:    MsgText{content},
	})
	if err != nil {
		return fmt.Errorf("send wework msg failed: %w", err)
	}
	resp, err := httpc.Post(s.BaseURL+key, "application/json", bytes.NewReader(msg))
	if err != nil {
		return fmt.Errorf("wechat work send robot message api error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
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
		return s.SendRobotMsg(s.ErrorKey, entry.Message+"\n\n"+entry.Caller.String()+"\n\n"+entry.Stack)
	}
	return nil
}
