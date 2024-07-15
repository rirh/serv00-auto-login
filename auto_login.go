package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

type Account struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Panel    string `json:"panel"`
}

var (
	telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID   = os.Getenv("TELEGRAM_CHAT_ID")
	message          = "serv00&ct8自动化脚本运行\n"
)

func formatToISO(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func delayTime(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func login(ctx context.Context, username, password, panel string) (bool, error) {
	serviceName := "ct8"
	if !contains(panel, "ct8") {
		serviceName = "serv00"
	}

	var loggedIn bool

	url := fmt.Sprintf("https://%s/login/?next=/", panel)
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.SendKeys(`#id_username`, username),
		chromedp.SendKeys(`#id_password`, password),
		chromedp.Click(`#submit`),
		chromedp.WaitVisible(`a[href="/logout/"]`),
		chromedp.Evaluate(`document.querySelector('a[href="/logout/"]') !== null`, &loggedIn),
	); err != nil {
		return false, fmt.Errorf("%s账号 %s 登录时出现错误: %v", serviceName, username, err)
	}

	return loggedIn, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func sendTelegramMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)
	payload := map[string]interface{}{
		"chat_id": telegramChatID,
		"text":    message,
		"reply_markup": map[string]interface{}{
			"inline_keyboard": [][]map[string]string{
				{
					{
						"text": "问题反馈❓",
						"url":  "https://t.me/yxjsjl",
					},
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("发送消息到Telegram时出错: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("发送消息到Telegram失败: %s", string(bodyBytes))
	}

	return nil
}

func main() {
	message = "serv00&ct8自动化脚本运行\n"

	accountsJSON := os.Getenv("ACCOUNTS_JSON")
	if accountsJSON == "" {
		log.Fatalf("环境变量 ACCOUNTS_JSON 为空")
	}

	var accounts []Account
	if err := json.Unmarshal([]byte(accountsJSON), &accounts); err != nil {
		log.Fatalf("解析 ACCOUNTS_JSON 环境变量时出错: %v", err)
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	for _, account := range accounts {
		serviceName := "ct8"
		if !contains(account.Panel, "ct8") {
			serviceName = "serv00"
		}

		isLoggedIn, err := login(ctx, account.Username, account.Password, account.Panel)
		if err != nil {
			message += fmt.Sprintf("%s账号 %s 登录失败，请检查%s账号和密码是否正确。\n", serviceName, account.Username, serviceName)
			log.Println(err)
		} else if isLoggedIn {
			nowUTC := formatToISO(time.Now().UTC())
			nowBeijing := formatToISO(time.Now().UTC().Add(8 * time.Hour))
			successMessage := fmt.Sprintf("%s账号 %s 于北京时间 %s（UTC时间 %s）登录成功！", serviceName, account.Username, nowBeijing, nowUTC)
			message += successMessage + "\n"
			log.Println(successMessage)
		} else {
			message += fmt.Sprintf("%s账号 %s 登录失败，请检查%s账号和密码是否正确。\n", serviceName, account.Username, serviceName)
			log.Printf("%s账号 %s 登录失败，请检查%s账号和密码是否正确。\n", serviceName, account.Username, serviceName)
		}

		delay := rand.Intn(7000) + 1000
		delayTime(delay)
	}

	message += "所有账号登录完成！"
	if err := sendTelegramMessage(message); err != nil {
		log.Println(err)
	}
}
