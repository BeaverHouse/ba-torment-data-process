# 데이터베이스 명세

블루아카이브 총력전 사이트에서 사용하는 데이터베이스 (이하 DB) 명세입니다.

## DB 정보

- DB Type: PostgreSQL
- Hosting: [Supabase](https://supabase.com)

## 테이블 목록

### named_users

총력전 영상과 유저 정보를 매핑합니다.

| 컬럼명      | 타입         | NULL     | 기본값            | 설명                                      |
| ----------- | ------------ | -------- | ----------------- | ----------------------------------------- |
| user_id     | int4         | NOT NULL | -                 | 유저 ID                                   |
| raid_id     | varchar(20)  | NULL     | -                 | 총력전 ID (채널 정보일 경우 NULL)         |
| description | varchar(200) | NOT NULL | -                 | 유저 설명                                 |
| youtube_url | varchar(200) | NOT NULL | -                 | 유튜브 영상 URL 또는 해당 유저의 채널 URL |
| created_at  | timestamp    | NOT NULL | CURRENT_TIMESTAMP | 생성일                                    |
| updated_at  | timestamp    | NULL     | -                 | 수정일                                    |
| deleted_at  | timestamp    | NULL     | -                 | 삭제일                                    |
| score       | int4         | NOT NULL | -                 | 점수 (채널 정보일 경우 0)                 |

**제약조건**

- UNIQUE (user_id, youtube_url)

### raids

현재 사이트에 제공되는 총력전 정보를 저장합니다.

| 컬럼명     | 타입         | NULL     | 기본값            | 설명                                               |
| ---------- | ------------ | -------- | ----------------- | -------------------------------------------------- |
| raid_id    | varchar(20)  | NOT NULL | -                 | 총력전 ID                                          |
| name       | varchar(200) | NOT NULL | -                 | 총력전 이름                                        |
| status     | varchar(20)  | NOT NULL | -                 | 총력전 데이터 처리 상태 (PENDING or COMPLETE)      |
| created_at | timestamp    | NOT NULL | CURRENT_TIMESTAMP | 생성일                                             |
| updated_at | timestamp    | NULL     | -                 | 수정일                                             |
| deleted_at | timestamp    | NULL     | -                 | 삭제일                                             |
| is_lunatic | bool         | NULL     | false             | 현재는 의미 없는 데이터                            |
| top_level  | varchar      | NULL     | -                 | 가장 높은 난이도 (L: 루나틱, T: 토먼트, I: 인세인) |

**제약조건**

- PRIMARY KEY (raid_id)

### students

학생 정보를 저장합니다.

| 컬럼명     | 타입        | NULL     | 기본값            | 설명      |
| ---------- | ----------- | -------- | ----------------- | --------- |
| student_id | varchar(20) | NOT NULL | -                 | 학생 ID   |
| name       | varchar(50) | NOT NULL | -                 | 학생 이름 |
| created_at | timestamp   | NOT NULL | CURRENT_TIMESTAMP | 생성일    |
| updated_at | timestamp   | NULL     | -                 | 수정일    |
| deleted_at | timestamp   | NULL     | -                 | 삭제일    |

**제약조건**

- PRIMARY KEY (student_id)
