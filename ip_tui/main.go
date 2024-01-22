package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type IpInfo struct {
	Query      string  `json:"query"`
	Status     string  `json:"status"`
	Message    string  `json:"message"`
	Country    string  `json:"country"`
	RegionName string  `json:"regionName"`
	City       string  `json:"city"`
	Zip        string  `json:"zip"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Timezone   string  `json:"timezone"`
	Isp        string  `json:"isp"`
}

func grabIp(ip string) (IpInfo, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%v", ip)
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(url)

	if err != nil {
		return IpInfo{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IpInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ipInfo IpInfo

	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		return IpInfo{}, fmt.Errorf("error decoding response body: %v", err)
	}
	return ipInfo, nil
}

func main() {
	var ipSearch string
	app := tview.NewApplication()
	pages := tview.NewPages()

	inputField := tview.NewInputField().
		SetLabel("Enter Ip Address: ").
		SetFieldTextColor(tcell.ColorBlack).
		SetChangedFunc(
			func(text string) {
				ipSearch = text
			},
		).
		SetDoneFunc(
			func(key tcell.Key) {
				ipInfo, e := grabIp(ipSearch)

				if e != nil {
					fmt.Printf("error occured with call: %v", e)
					os.Exit(1)
				}

				if ipInfo.Status == "success" {
					pages.AddPage(
						"results", tview.NewList().
							AddItem("IP", ipInfo.Query, ' ', nil).
							AddItem("Country", ipInfo.Country, ' ', nil).
							AddItem("Region", ipInfo.RegionName, ' ', nil).
							AddItem("City", ipInfo.City, ' ', nil).
							AddItem("Zip", ipInfo.Zip, ' ', nil).
							AddItem(
								"Lat + Lon",
								fmt.Sprintf(
									"(%v, %v)",
									ipInfo.Lat, ipInfo.Lon,
								), ' ', nil,
							).
							AddItem("Timezone", ipInfo.Timezone, ' ', nil).
							AddItem("ISP", ipInfo.Isp, ' ', nil), true, true,
					)
				} else {
					pages.AddPage(
						"results", tview.NewTextView().
							SetDynamicColors(true).
							SetText(fmt.Sprintf("Search for %v failed\n Reason: %v", ipInfo.Query, ipInfo.Message)),
						true, true,
					)
				}
			},
		)

	if err := app.SetRoot(
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(pages, 0, 2, false).
			AddItem(inputField, 2, 1, true), true,
	).Run(); err != nil {
		panic(err)
	}
}
