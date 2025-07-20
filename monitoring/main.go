package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
	seoulRPC     = "http://3.35.221.3:26657"
	tokyoRPC     = "http://52.196.214.161:26657"
	seoulValCons = "84F5CA9595D151C75915A2EDDD33CA09884F4AD1"
	tokyoValCons = "1DFC650985DB21E7EE21AE9F9D8EBE9A8C7C4AD6"
	webhookURL   = "https://discord.com/api/webhooks/..." // 실제 Webhook 주소 입력
)

func sendDiscordAlert(message string) {
	payload := map[string]string{"content": message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Println("디스코드 JSON 변환 실패:", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("⚠ 디스코드 전송 실패:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		log.Println("디스코드 응답 코드:", resp.StatusCode)
	}
}

type syncState struct {
	lastHeight int64
	lastCatchingTime time.Time
}

func monitorNode(name string, rpcURL string, valCons string, state *syncState) {
	client, err := tmhttp.New(rpcURL)
	if err != nil {
		log.Fatalf("%s 노드 연결 실패: %v", name, err)
	}

	for {
		status, err := client.Status(context.Background())
		if err != nil {
			sendDiscordAlert(fmt.Sprintf("%s 노드에 연결 실패: %v", name, err))
			time.Sleep(2 * time.Second)
			continue
		}

		syncInfo := status.SyncInfo
		height := syncInfo.LatestBlockHeight

		// 동기화 상태 감지
		if syncInfo.CatchingUp {
			if height == state.lastHeight {
				// height 멈춤 + catching 상태 유지
				if state.lastCatchingTime.IsZero() {
					state.lastCatchingTime = time.Now()
				} else if time.Since(state.lastCatchingTime) > 2*time.Second {
					netInfo, err := client.NetInfo(context.Background())
					if err == nil && len(netInfo.Peers) == 0 {
						sendDiscordAlert(fmt.Sprintf("노드 동기화 실패: 2초 이상 CatchingUp + 블록 미증가 + 피어 없음 (%s)", name))
					}
				}
			} else {
				state.lastHeight = height
				state.lastCatchingTime = time.Now()
			}
		} else {
			state.lastCatchingTime = time.Time{}
			state.lastHeight = height
		}

		// 서명 누락 감지
		block, err := client.Block(context.Background(), nil)
		if err == nil {
			signatures := block.Block.LastCommit.Signatures
			found := false
			for _, sig := range signatures {
				if sig.ValidatorAddress.String() == valCons && sig.BlockIDFlag == 2 {
					found = true
					break
				}
			}
			if !found {
				msg := fmt.Sprintf("%s 블록 %d에서 서명 누락", name, block.Block.Height)
				sendDiscordAlert(msg)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	go monitorNode("Seoul", seoulRPC, seoulValCons, &syncState{})
	go monitorNode("Tokyo", tokyoRPC, tokyoValCons, &syncState{})

	select {} // 무한 대기
}
