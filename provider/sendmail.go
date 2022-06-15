package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/library"
	"os"
	"strings"
	"time"
)

const MailLogFile = "mail"

type MailLog struct {
	CreatedTime int64  `json:"created_time"`
	Subject     string `json:"subject"`
	Status      string `json:"status"`
}

func GetLastSendmailList() ([]*MailLog, error) {
	var mailLogs []*MailLog
	//获取20条数据
	filePath := fmt.Sprintf("%scache/%s.log", config.ExecPath, MailLogFile)
	logFile, err := os.Open(filePath)
	if nil != err {
		//打开失败
		return mailLogs, nil
	}
	defer logFile.Close()

	line := int64(1)
	cursor := int64(0)
	stat, err := logFile.Stat()
	fileSize := stat.Size()
	tmp := ""
	for {
		cursor -= 1
		logFile.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		logFile.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			//跳到一个新行，清空
			line++
			//解析
			if tmp != "" {
				var mailLog MailLog
				err := json.Unmarshal([]byte(tmp), &mailLog)
				if err == nil {
					mailLogs = append(mailLogs, &mailLog)
				}
			}
			tmp = ""
		}

		tmp = fmt.Sprintf("%s%s", string(char), tmp)

		if cursor == -fileSize {
			// stop if we are at the beginning
			break
		}
		if line == 100 {
			break
		}
	}
	//解析最后一条
	if tmp != "" {
		var mailLog MailLog
		err := json.Unmarshal([]byte(tmp), &mailLog)
		if err == nil {
			mailLogs = append(mailLogs, &mailLog)
		}
	}

	return mailLogs, nil
}

func SendMail(subject, content string) error {
	setting := config.JsonData.PluginSendmail
	port := setting.Port
	if port == 0 {
		//默认使用25端口
		port = 25
	}
	if setting.UseSSL == 1 && port == 25 {
		//如果使用ssl，设置了25端口，则使用465
		port = 465
	}

	if setting.Account == "" {
		//成功配置，则跳过
		return errors.New(config.Lang("请配置发件人信息"))
	}

	//开始发送
	email := library.NewEMail(`{"port":25}`)
	email.From = setting.Account
	email.Host = setting.Server
	email.Port = setting.Port
	email.Username = setting.Account
	if setting.UseSSL == 1 {
		email.Secure = "SSL"
	}
	email.Password = setting.Password

	var recipients []string
	if setting.Recipient != "" {
		tmp := strings.Split(setting.Recipient, ",")
		for _, v := range tmp {
			v = strings.TrimSpace(v)
			if v != "" {
				recipients = append(recipients, v)
			}
		}
	}
	if len(recipients) == 0 {
		recipients = append(recipients, setting.Account)
	}

	email.To = recipients
	email.Subject = subject
	email.Text = content

	if err := email.Send(); err != nil {
		logMailError(subject, err.Error())
		return err
	}
	logMailError(subject, config.Lang("发送成功"))
	return nil
}

func logMailError(subject, status string) {
	mailLog := MailLog{
		CreatedTime: time.Now().Unix(),
		Subject:     subject,
		Status:      status,
	}

	content, err := json.Marshal(mailLog)

	if err == nil {
		library.DebugLog(MailLogFile, string(content))
	}
}
