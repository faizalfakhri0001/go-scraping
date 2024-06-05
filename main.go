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

	// namaProvinsi := "ACEH"
	listKabupaten, err := getKabupatenByProvince(f, 11)
	if err != nil {
		fmt.Println(err)
		return
	}

	sheet := "HALF"
	rows, err := f.GetRows(sheet)
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 2)

	before := 1
	for idx, kab := range *listKabupaten {
		jumlah, err := strconv.Atoi(kab[3])
		if err != nil {
			fmt.Println(err)
			break

		}
		listKelurahanPerKabupaten := rows[before : before+jumlah]

		wg.Add(1)
		sem <- struct{}{}

		go func(list [][]string, kab_name string, no int) {
			defer wg.Done()
			defer func() { <-sem }()

			newFile := excelize.NewFile()
			defer func() {
				if err := newFile.Close(); err != nil {
					fmt.Println(err)
				}
			}()
			new_sheet := "Hasil"
			index, err := newFile.NewSheet(new_sheet)
			if err != nil {
				fmt.Println(err)
				return
			}
			newFile.SetActiveSheet(index)

			ctx1, closeBrowser := chromedp.NewContext(allocCtx)
			defer closeBrowser()

			fmt.Printf("%d.Nama Kabupaten: %s\n", no, kab_name)
			fmt.Printf("  Jumlah Kelurahan: %d\n", len(listKelurahanPerKabupaten))
			fmt.Printf("  Kelurahan Pertama: %s\n", listKelurahanPerKabupaten[0][0])
			fmt.Printf("  Kelurahan Terakhir: %s\n\n", listKelurahanPerKabupaten[len(listKelurahanPerKabupaten)-1][0])

			for idxKel, kelurahan := range list {
				url := fmt.Sprintf("https://kawalpemilu.org/h/%s", kelurahan[4])
				var nodes []*cdp.Node
				if err := chromedp.Run(
					ctx1,
					chromedp.Navigate(url),
					chromedp.WaitVisible(".tps-item", chromedp.ByQueryAll),
					chromedp.Nodes(".tps-item", &nodes, chromedp.ByQueryAll),
				); err != nil {
					fmt.Println("TPS tidak ditemukan:", err)
				}

				var dpt, hs string
				var listDPT []map[string]int
				var listHS []map[string][]int
				totalDPT := 0

				for indexNodes, node := range nodes {
					chromedp.Run(ctx1,
						chromedp.WaitVisible(".tps-dpt", chromedp.ByQuery, chromedp.FromNode(node)),
						chromedp.WaitVisible("span.tps-result", chromedp.ByQuery, chromedp.FromNode(node)),
						chromedp.Text(".tps-dpt", &dpt, chromedp.ByQuery, chromedp.FromNode(node)),
						chromedp.Text("span.tps-result", &hs, chromedp.ByQuery, chromedp.FromNode(node)),
					)

					re := regexp.MustCompile(`\d+`)
					match := re.FindString(dpt)
					number, err := strconv.Atoi(match)
					if err != nil {
						fmt.Printf("\tTPS %d: DPT tidak ditemukan\n", indexNodes+1)
					} else {
						// fmt.Printf("\tTPS %d: DPT ditemukan\n", indexNodes+1)
						totalDPT += number
						tpsPos := fmt.Sprintf("TPS%d", indexNodes+1)
						listDPT = append(listDPT, map[string]int{tpsPos: number})

						newMap := make(map[string][]int)
						HSpaslon := extractValues(hs)
						newMap[tpsPos] = HSpaslon
						listHS = append(listHS, newMap)

						// fmt.Printf("\t  HS Paslon: %v \n", listHS)
						// fmt.Printf("\t  List DPT: %v \n", listDPT)
						// fmt.Printf("\t  Total DPT: %v \n", totalDPT)
					}
				}
				output := ScripeOut{
					HS:            listHS,
					DPT:           listDPT,
					Total:         totalDPT,
					NamaKelurahan: kelurahan[0],
					IDKelurahan:   kelurahan[4],
					Line:          idxKel + 1,
				}

				setValuesInCell(newFile, &output, new_sheet)

				if idxKel == 10 {
					break
				}
			}

			if err := newFile.SaveAs(fmt.Sprintf("outputs/%s.xlsx", kab_name)); err != nil {
				fmt.Println(err)
			}
		}(listKelurahanPerKabupaten, kab[0], idx+1)
		before += jumlah
		if idx == 3 {
			break
		}
	}
	wg.Wait()
}

func setValuesInCell(f *excelize.File, datas *ScripeOut, new_sheet string) {
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

func getKabupatenByProvince(f *excelize.File, province_id int) (*[][]string, error) {
	var filteredRows [][]string

	sheet := "Kabupaten"
	rows, err := f.GetRows(sheet)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, row := range rows {
		id, err := strconv.Atoi(row[2])
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if id == province_id {
			filteredRows = append(filteredRows, row)
		} else if id > province_id {
			break
		}
	}

	return &filteredRows, nil
}
