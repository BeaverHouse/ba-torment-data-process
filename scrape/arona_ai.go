package scrape

import (
	"ba-torment-data-process/app/logic"
	"context"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func GetDataFromAronaAI(seasonString string) (string, error) {
	path := "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"
	if runtime.GOOS == "darwin" {
		path = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	parentContext, parentContextCancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.ExecPath(path),
	)
	defer parentContextCancel()

	// create context
	ctx, cancel := chromedp.NewContext(parentContext)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	season, category, err := logic.SplitSeasonString(seasonString)
	if err != nil {
		return "", err
	}
	// S와 3S를 제거하고 숫자로 변환
	if strings.HasPrefix(season, "3S") {
		season = season[2:]
	} else if strings.HasPrefix(season, "S") {
		season = season[1:]
	}

	url := "https://arona.ai/raidreport"
	if category != 0 {
		url = "https://arona.ai/eraidreport"
	}

	var tValue string

	log.Printf("Navigating to %s", url)
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		log.Printf("Navigation failed: %v", err)
		return "", err
	}

	log.Printf("Waiting for combobox and opening season selector")
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`//div[@role="combobox"]`, chromedp.BySearch),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`//div[@role="combobox"]`, chromedp.NodeVisible),
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		log.Printf("Failed to open season selector: %v", err)
		return "", err
	}

	log.Printf("Checking if season %s exists", season)
	var seasonExists bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('li[data-value="`+season+`"]') !== null`, &seasonExists),
	)
	if err != nil {
		log.Printf("Failed to check season existence: %v", err)
		return "", err
	}

	// Season이 존재하지 않으면 로그 출력 후 종료
	if !seasonExists {
		log.Printf("Season %s not found in arona.ai", season)
		return "", nil
	}

	log.Printf("Selecting season %s", season)
	err = chromedp.Run(ctx,
		chromedp.Click(`//li[@data-value="`+season+`"]`, chromedp.NodeVisible),
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		log.Printf("Failed to select season: %v", err)
		return "", err
	}

	// 1-1. Category가 0이 아닐 경우, 알맞은 대결전을 선택해야 합니다.
	// 맨 마지막 MuiToggleButtonGroup-root div에 있는 category번째 button을 클릭합니다.
	if category != 0 {
		err = chromedp.Run(ctx,
			chromedp.WaitVisible(`//div[contains(@class,"MuiToggleButtonGroup-root")]`),
			chromedp.Sleep(1*time.Second),
			chromedp.ScrollIntoView(`(//div[contains(@class,"MuiToggleButtonGroup-root")])[last()]`),
			chromedp.Sleep(1*time.Second), // 스크롤 완료 대기
			// 맨 마지막 MuiToggleButtonGroup-root div에 있는 category번째 button을 클릭합니다.
			chromedp.Click(`(//div[contains(@class,"MuiToggleButtonGroup-root")])[last()]//button[`+strconv.Itoa(category)+`]`, chromedp.NodeVisible),
			chromedp.Sleep(1*time.Second),
		)
		if err != nil {
			return "", err
		}
	}

	err = chromedp.Run(ctx,

		// 2. "직접 검색" 버튼이 보일 때까지 스크롤하고 클릭합니다.
		chromedp.WaitVisible(`//div[text()="직접 검색"]`),
		chromedp.Sleep(1*time.Second),
		chromedp.ScrollIntoView(`//div[text()="직접 검색"]`),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`//div[text()="직접 검색"]`, chromedp.NodeVisible),
		chromedp.Sleep(1*time.Second), // UI 업데이트 대기

		// JavaScript 주입: JSON.parse를 가로채서 't' 값을 전역 변수에 저장
		chromedp.Evaluate(`(() => {
			const JSON_parse_original = JSON.parse;
			JSON.parse = (t, r) => {
				window.aronaDataT = t; // 't' 값을 전역 변수에 저장
				JSON.parse = JSON_parse_original; // 원래 JSON.parse 함수 복원
				return JSON_parse_original(t, r); // 원래 JSON.parse 호출
			};
		})();`, nil),
		chromedp.Sleep(1*time.Second),

		// 3. role="combobox"인 5번째 div를 클릭합니다.
		chromedp.WaitVisible(`(//div[@role="combobox"])[5]`),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`(//div[@role="combobox"])[5]`, chromedp.NodeVisible),
		chromedp.Sleep(1*time.Second),

		// 4. data-value="20000"인 li를 클릭합니다.
		chromedp.WaitVisible(`//li[@data-value="20000"]`),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`//li[@data-value="20000"]`, chromedp.NodeVisible),

		// 5. 데이터가 로드되기를 기다립니다.
		chromedp.Sleep(3*time.Second),

		// 6. 주입된 JavaScript를 통해 전역 변수에 저장된 't' 값 추출
		chromedp.Evaluate(`window.aronaDataT`, &tValue),
	)
	if err != nil {
		return "", err
	}

	log.Printf("Successfully retrieved data from arona.ai")

	return tValue, nil
}
