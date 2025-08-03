package scrape

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func GetDataFromAronaAI() (string, error) {
	parentContext, parentContextCancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.ExecPath(`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`),
	)
	defer parentContextCancel()

	// create context
	ctx, cancel := chromedp.NewContext(parentContext)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var tValue string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://arona.ai/raidreport`),

		// 1. "직접 검색" 버튼이 보일 때까지 스크롤하고 클릭합니다.
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

		// 2. role="combobox"인 5번째 div를 클릭합니다.
		chromedp.WaitVisible(`(//div[@role="combobox"])[5]`),
		chromedp.Click(`(//div[@role="combobox"])[5]`, chromedp.NodeVisible),

		// 3. data-value="20000"인 li를 클릭합니다.
		chromedp.WaitVisible(`//li[@data-value="20000"]`),
		chromedp.Click(`//li[@data-value="20000"]`, chromedp.NodeVisible),

		// 4. 데이터가 로드되기를 기다립니다.
		chromedp.Sleep(3*time.Second),

		// 5. 주입된 JavaScript를 통해 전역 변수에 저장된 't' 값 추출
		chromedp.Evaluate(`window.aronaDataT`, &tValue),
	)
	if err != nil {
		return "", err
	}

	log.Printf("Successfully retrieved 't' variable: %s", tValue)

	return tValue, nil
}
