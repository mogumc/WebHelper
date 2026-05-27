package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func flashtime(a *App) {
	go func() {
		for {
			now := time.Now().Format(time.DateTime)
			runtime.EventsEmit(a.ctx, "time", now)
			time.Sleep(time.Second)
		}
	}()
}

func getJSONString() string {
	datas := map[string]interface{}{
		"code":    200,
		"message": "ok",
		"data": []map[string]interface{}{
			{
				"id":    1,
				"name":  "apps",
				"isdir": true,
				"size":  0,
			},
			{
				"id":    2,
				"name":  "docs",
				"isdir": true,
				"size":  0,
			},
			{
				"id":    3,
				"name":  "readme.txt",
				"isdir": false,
				"size":  2048,
			},
		},
	}

	jsonBytes, err := json.Marshal(datas)
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}

type SystemInfo struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	NumCPU      int    `json:"num_cpu"`
	Hostname    string `json:"hostname"`
	GoVer       string `json:"go_ver"`
	Time        string `json:"time"`
	ProcessName string `json:"process_name"`
}
