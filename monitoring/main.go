package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    tmhttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
    seoulRPC     = "http://3.35.221.3:26657"
    tokyoRPC     = "http://52.196.214.161:26657"
    seoulValCons = "84F5CA9595D151C75915A2EDDD33CA09884F4AD1"
    tokyoValCons = "1DFC650985DB21E7EE21AE9F9D8EBE9A8C7C4AD6"
    webhookURL   = "https://discord.com/api/webhooks/1396388049537073173/eRCtaFBUux4HGJKrVEDUvg-BbWUWriJYTpqcCV9_h9g7L4L388684mGKFBznsR7x1WtG"
)

// Discord로 알림
func sendDiscordAlert(msg string) {
    payload := map[string]string{"content": msg}
    data, err := json.Marshal(payload)
    if err != nil {
        fmt.Printf("JSON 변환 에러: %v\n", err)
        return
    }

    resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
    if err != nil {
        fmt.Printf("Discord 전송 실패: %v\n", err)
        return
    }

    resp.Body.Close()
}

// 노드 상태 추적 구조체
type State struct {
    lastH int64
    lastT time.Time
}

func monitorNode(name, rpcURL, valCons string, st *State) {
    cli, err := tmhttp.New(rpcURL)
    if err != nil {
        fmt.Printf("%s 연결 실패! err=%v\n", name, err)
        return
    }

    for i := 0; ; i++ {
        // 1) Status 가져오기
        stat, err := cli.Status(context.Background())
        if err != nil {
            sendDiscordAlert(fmt.Sprintf("%s Status 에러: %v", name, err))
            time.Sleep(2 * time.Second)
            continue
        }
        h := stat.SyncInfo.LatestBlockHeight

        // 2) 동기화 멈춤 감지
        if stat.SyncInfo.CatchingUp {
            if h == st.lastH {
                if st.lastT.IsZero() {
                    st.lastT = time.Now()
                } else if time.Since(st.lastT) > 2*time.Second {
                    // 피어도 없으면 경고
                    netInfo, _ := cli.NetInfo(context.Background())
                    if len(netInfo.Peers) == 0 {
                        sendDiscordAlert(fmt.Sprintf("%s 노드 동기화 중단! height=%d", name, h))
                    }
                }
            } else {
                // 블록이 증가했으면 리셋
                st.lastH = h
                st.lastT = time.Now()
            }
        } else {
            // 정상 상태면 초기화
            st.lastH = h
            st.lastT = time.Time{}
        }

        // 3) 서명 누락 체크
        blockRes, err := cli.Block(context.Background(), nil)
        if err == nil {
            sigs := blockRes.Block.LastCommit.Signatures
            ok := false
            for _, s := range sigs {
                // BlockIDFlag == 2 은 정상 서명
                if s.ValidatorAddress.String() == valCons && s.BlockIDFlag == 2 {
                    ok = true
                    break
                }
            }
            if !ok {
                sendDiscordAlert(fmt.Sprintf("%s: 블록 %d 서명 누락!", name, blockRes.Block.Height))
            }
        }

        fmt.Printf("[%s] #%d loop done (height=%d)\n", name, i, h)
        time.Sleep(500 * time.Millisecond)
    }
}

func main() {
    seoulState := &State{}
    tokyoState := &State{}

    go monitorNode("Seoul", seoulRPC, seoulValCons, seoulState)
    go monitorNode("Tokyo", tokyoRPC, tokyoValCons, tokyoState)

    select {}
}