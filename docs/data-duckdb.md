# DuckDB Data 명세

https://www.plana-stats.com 에서 사용하는 DuckDB 파일을 다운로드받아 사용합니다.

## 데이터 형식

1. `complete_runs` 테이블
   - `crunid`: 완료된 총력전/대결전 도전 ID
   - `point`: 총력전/대결전 점수
     - 대결전일 경우, `ArmorTypeMapping`에 있는 장갑 타입이 Column 이름에 포함됩니다.
   - 그 외에는 보지 않아도 됩니다.
2. `runs*` 테이블
   - 대결전일 경우, `ArmorTypeMapping`에 있는 장갑 타입이 테이블 이름에 포함됩니다.
   - `crunid`: 완료된 총력전/대결전 도전 ID
   - `runid`: 개별 도전 ID
   - `runcount`: 도전 회차 (ex. 2일 경우 2파티를 의미합니다.)
3. `students*` 테이블
   - 대결전일 경우, `ArmorTypeMapping`에 있는 장갑 타입이 테이블 이름에 포함됩니다.
   - `runid`: 개별 도전 ID
   - `sid`: 학생 ID
   - `build`: 학생 성급을 나타냅니다. 자세한 매핑은 `WeaponStarMapping`을 참고하세요.
     - `three`, `four` 등은 전용무기가 없는 n성 캐릭터입니다.
     - `UE30`, `UE40` 등은 전용무기가 있는 5성 캐릭터입니다. 예를 들어, UE40은 전무 2성입니다. (전무 2성의 최대 무기 레벨이 40이기 때문입니다.)
   - `slot`: 0~3은 각각 1~4번째 스트라이커, 4는 스페셜 학생입니다.
   - `assist`: 조력자 여부입니다.
   - `level`: 학생의 레벨입니다. 보통은 최대 레벨입니다.
