package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/gorilla/websocket"
    "github.com/tencent-connect/botgo"
    "github.com/tencent-connect/botgo/dto"
    "github.com/tencent-connect/botgo/openapi"
    "github.com/tencent-connect/botgo/token"
    yaml "gopkg.in/yaml.v2"
)

type Config struct {
    AppID uint64 `yaml:"appid"`
    Token string `yaml:"token"`
}

var config Config
var api openapi.OpenAPI
var ctx context.Context

func init() {
    content, err := ioutil.ReadFile("config.yaml")//我创造了一个yaml文件里面放了我的机器人的id和token
    if err != nil {
        log.Fatalf("读取配置文件出错: %v", err)
    }
    if err := yaml.Unmarshal(content, &config); err != nil {
        log.Fatalf("解析配置文件出错: %v", err)
    }
}

func atMessageEventHandler(message string, channelID string, msgID string) {
    if strings.HasSuffix(message, "hello") {
        api.PostMessage(ctx, channelID, &dto.MessageToCreate{MsgID: msgID, Content: "你好"})
    }
}

func sendReadyMessage(channelID string) {
    msg := &dto.MessageToCreate{
        Content: "面试官您好，您的专属机器人已准备就绪！",
    }
    _, err := api.PostMessage(ctx, channelID, msg)
    if err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func connectAndListen() {
    wsURL := "wss://api.sgroup.qq.com/websocket"
    header := http.Header{}
    header.Add("Authorization", "Bot "+fmt.Sprintf("%d.%s", config.AppID, config.Token))

    for {
        wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
        if err != nil {
            log.Printf("WebSocket连接失败: %v，将在5秒后重试", err)
            time.Sleep(5 * time.Second)
            continue
        }
        log.Println("WebSocket连接成功")

        sendReadyMessage("640682838") //这里是我的频道ID

        defer wsConn.Close()

        wsConn.SetPingHandler(func(appData string) error {
            deadline := time.Now().Add(5 * time.Second)
            return wsConn.WriteControl(websocket.PongMessage, []byte(appData), deadline)
        })

        go func() {
            for {
                time.Sleep(30 * time.Second)
                deadline := time.Now().Add(5 * time.Second)
                if err := wsConn.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
                    log.Printf("发送心跳失败: %v", err)
                    return
                }
            }
        }()

        for {
            _, message, err := wsConn.ReadMessage()
            if err != nil {
                log.Printf("读取消息错误: %v，将尝试重新连接", err)
                break
            }

            atMessageEventHandler(string(message), "640682838", "08e9bb9dd5fc93cbf51910d696c0b102382e48ed909fb006")//这里填的我的频道的id和msg
        }

        time.Sleep(5 * time.Second)
    }
}

func main() {
    botToken := token.BotToken(config.AppID, config.Token)
    api = botgo.NewOpenAPI(botToken).WithTimeout(3 * time.Second)
    ctx = context.Background()

    connectAndListen()
}
