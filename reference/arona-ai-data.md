# Arona AI Data 명세

> [!NOTE]
> This document is written in Korean.

[arona.ai](https://arona.ai)에서 제공하는 총력전 데이터 명세입니다.

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
