package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/xuri/excelize/v2"
)

type DPT = map[string]int
type HS = map[string][]int

type ScripeOut struct {
	HS            []HS
	DPT           []DPT
	Total         int
	NamaKelurahan string
	IDKelurahan   string
	Line          int
}

func main() {
	f, err := excelize.OpenFile("data2.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheet := "HALF"
	rows, err := f.GetRows(sheet)
	if err != nil {
		fmt.Println(err)
		return
	}

	new_f := excelize.NewFile()
	defer func() {
		if err := new_f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	index, err := new_f.NewSheet("Hasil")
	if err != nil {
		fmt.Println(err)
		return
	}
	new_f.SetActiveSheet(index)

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// new browser, first tab
	ctx1, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(ctx1); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for idx, row := range rows {
		if idx == 0 {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			url := fmt.Sprintf("https://kawalpemilu.org/h/%s", row[4])

			tabCtx, closeTab := chromedp.NewContext(ctx1)
			defer closeTab()

			var nodes []*cdp.Node
			if err := chromedp.Run(
				tabCtx,
				chromedp.Navigate(url),
				chromedp.Nodes(".tps-item", &nodes, chromedp.ByQueryAll),
			); err != nil {
				panic(err)
			}

			var dpt, hs string
			var listDPT []map[string]int
			var listHS []map[string][]int
			totalDPT := 0

			fmt.Printf("ID Kelurahan: %s\n", row[4])
			for indexNodes, node := range nodes {
				chromedp.Run(tabCtx,
					chromedp.Text(".tps-dpt", &dpt, chromedp.ByQuery, chromedp.FromNode(node)),
					chromedp.Text("span.tps-result", &hs, chromedp.ByQuery, chromedp.FromNode(node)),
				)

				re := regexp.MustCompile(`\d+`)
				match := re.FindString(dpt)
				number, err := strconv.Atoi(match)
				if err != nil {
					fmt.Printf("\tTPS %d: DPT tidak ditemukan\n", indexNodes+1)
				} else {
					fmt.Printf("\tTPS %d: DPT ditemukan\n", indexNodes+1)
					totalDPT += number
					tpsPos := fmt.Sprintf("TPS%d", indexNodes+1)
					listDPT = append(listDPT, map[string]int{tpsPos: number})

					newMap := make(map[string][]int)
					HSpaslon := extractValues(hs)
					newMap[tpsPos] = HSpaslon
					listHS = append(listHS, newMap)

					fmt.Printf("\t\tHS Paslon: %v \n", listHS)
					fmt.Printf("\t\tList DPT: %v \n", listDPT)
					fmt.Printf("\t\tTotal DPT: %v \n", totalDPT)

					output := ScripeOut{
						HS:            listHS,
						DPT:           listDPT,
						Total:         totalDPT,
						NamaKelurahan: row[0],
						IDKelurahan:   row[4],
						Line:          idx + 1,
					}

					setValuesInCell(new_f, &output)
				}
			}
		}()

		if idx == 10 {
			break
		}
	}

	if err := new_f.SaveAs("Results.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func setValuesInCell(f *excelize.File, datas *ScripeOut) {
	new_sheet := "Result"
	id, err := strconv.Atoi(datas.IDKelurahan)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	f.SetCellValue(new_sheet, fmt.Sprintf("A%d", datas.Line), datas.NamaKelurahan)
	f.SetCellValue(new_sheet, fmt.Sprintf("B%d", datas.Line), id)
	f.SetCellValue(new_sheet, fmt.Sprintf("C%d", datas.Line), datas.Total)
	f.SetCellValue(new_sheet, fmt.Sprintf("D%d", datas.Line), datas.HS)

	colNumber := 5
	for i, dpt := range datas.DPT {
		f.SetCellValue(new_sheet, fmt.Sprintf("%s%d", string(rune(64+colNumber)), datas.Line), dpt[fmt.Sprintf("TPS%d", i+1)])
		colNumber += 1
	}
}

func extractValues(data string) []int {
	lines := strings.Split(data, "\n")
	var values []int
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		values = append(values, intValue)
	}
	return values
}
