# Arona AI Data 명세

> [!NOTE]
> This document is written in Korean.

[arona.ai](https://arona.ai)의 총력전 데이터 명세입니다.

## 데이터 형식

```python
json_data = {
    "f": [], # 무시
    "d": [
        {
            "r": 1, # 등수
            "s": 1, # 총력전 점수
            "t": [  # 파티 정보 (여러 개일 수 있음)
                {
                    "m": [ # 스트라이커 (전열) 정보, length: 4
                        {
                            "id": 10098,        # 캐릭터 ID (예시는 타카하시 호시노(무장))
                            "star": 5,          # 캐릭터 성급
                            "level": 90,        # 캐릭터 레벨
                            "hasWeapon": true,  # 전용 무기 보유 여부
                            "isAssist": false,  # 조력자 여부
                            "weaponStar": 3,    # 전용 무기 성급
                            "isMulligan": false # 시작 커맨드 지정 여부
                        }
                    ],
                    "s": [ # 서포터 (후열) 정보, length: 2
                        # 서포터 정보 구조는 스트라이커 정보와 동일
                    ]
                }
            ]
        }
    ]
}
```

## 주의사항

2025.07월 이후 Arona AI의 데이터를 직접적으로 가져올 수 없게 되었습니다.  
꾸준히 업데이트가 이루어지고 암호화되고 있어 자동화를 통해 데이터를 가져오기 어렵다고 판단되고,  
현재는 수동으로 데이터를 가져오는 방식을 채택하였습니다.

1. 브라우저에서 Arona AI의 리포트에 접속한 후, 개발자 도구를 활성화합니다.
2. `raidreport-*.js` 또는 `eraidreport-*.js` 파일에서 Hydration 직전 시점을 찾습니다.
   - `isMulligan` 필드로 검색하여 찾는 것이 편합니다.
3. 중단점을 설정하고 디버깅을 통해 JSON 데이터를 얻습니다.
