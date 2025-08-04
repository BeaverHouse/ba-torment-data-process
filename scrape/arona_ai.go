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
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
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
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),

		// 1. 알맞은 시즌을 선택합니다. (seasonNum + . 으로 시작하는 div 클릭 (ex. 79. 으로 시작하는))
		chromedp.WaitVisible(`//div[@role="combobox"]`),
		chromedp.Click(`//div[@role="combobox"]`, chromedp.NodeVisible),
		chromedp.Click(`//li[@data-value="`+season+`"]`, chromedp.NodeVisible),
	)
	if err != nil {
		return "", err
	}

	// 1-1. Category가 0이 아닐 경우, 알맞은 대결전을 선택해야 합니다.
	// 맨 마지막 MuiToggleButtonGroup-root div에 있는 category번째 button을 클릭합니다.
	if category != 0 {
		err = chromedp.Run(ctx,
			chromedp.WaitVisible(`//div[contains(@class,"MuiToggleButtonGroup-root")]`),
			chromedp.ScrollIntoView(`(//div[contains(@class,"MuiToggleButtonGroup-root")])[last()]`),
			chromedp.Sleep(500*time.Millisecond), // 스크롤 완료 대기
			// 맨 마지막 MuiToggleButtonGroup-root div에 있는 category번째 button을 클릭합니다.
			chromedp.Click(`(//div[contains(@class,"MuiToggleButtonGroup-root")])[last()]//button[`+strconv.Itoa(category)+`]`, chromedp.NodeVisible),
		)
		if err != nil {
			return "", err
		}
	}

	err = chromedp.Run(ctx,

		// 2. "직접 검색" 버튼이 보일 때까지 스크롤하고 클릭합니다.
		chromedp.WaitVisible(`//div[text()="직접 검색"]`),
		chromedp.ScrollIntoView(`//div[text()="직접 검색"]`),
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

		// 3. role="combobox"인 5번째 div를 클릭합니다.
		chromedp.WaitVisible(`(//div[@role="combobox"])[5]`),
		chromedp.Click(`(//div[@role="combobox"])[5]`, chromedp.NodeVisible),

		// 4. data-value="20000"인 li를 클릭합니다.
		chromedp.WaitVisible(`//li[@data-value="20000"]`),
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
