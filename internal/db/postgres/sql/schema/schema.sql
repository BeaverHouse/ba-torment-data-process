CREATE SCHEMA IF NOT EXISTS batorment_v3;

SET timezone = 'Asia/Seoul';

CREATE TYPE top_level AS ENUM ('I', 'T', 'L');
CREATE TYPE analysis_type AS ENUM ('ai', 'user');

CREATE TABLE batorment_v3.students (
    student_id INTEGER NOT NULL PRIMARY KEY,
    name_ko VARCHAR(50) NOT NULL,
    name_ja VARCHAR(50) NOT NULL,
    search_keyword VARCHAR(50)[] NOT NULL,
    detail JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    name_en VARCHAR(50) NOT NULL DEFAULT '',
    name_zh VARCHAR(50) NOT NULL DEFAULT ''
);

CREATE TABLE batorment_v3.presents (
    present_id INTEGER NOT NULL PRIMARY KEY,
    name_ko VARCHAR(50) NOT NULL,
    rarity VARCHAR(10) NOT NULL,
    tags VARCHAR(50)[] NOT NULL,
    exp_value INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE batorment_v3.i18n (
    category VARCHAR(20) NOT NULL,
    key VARCHAR(50) NOT NULL,
    name_ko VARCHAR(100) NOT NULL,
    name_ja VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    name_zh VARCHAR(100) NOT NULL DEFAULT '',
    PRIMARY KEY (category, key)
);

CREATE TABLE batorment_v3.contents (
    content_id VARCHAR(10) NOT NULL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    top_level top_level NOT NULL DEFAULT 'T',
    start_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    post_discovered_at TIMESTAMPTZ,
    ko_start_date TIMESTAMPTZ
);

CREATE TABLE batorment_v3.youtube_analysis (
    id SERIAL PRIMARY KEY,
    video_id VARCHAR(255) NOT NULL,
    raid_id VARCHAR(10) NOT NULL,
    analysis_result JSONB NOT NULL,
    analysis_type analysis_type NOT NULL,
    version INTEGER NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE batorment_v3.youtube_metadata (
    id SERIAL PRIMARY KEY,
    video_id VARCHAR(255) NOT NULL,
    channel_url VARCHAR(500) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    raid_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    platform VARCHAR(20) NOT NULL DEFAULT 'youtube',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    CONSTRAINT youtube_metadata_video_id_raid_id_unique UNIQUE (video_id, raid_id),
    CONSTRAINT youtube_metadata_platform_check CHECK (platform IN ('youtube', 'bilibili'))
);

CREATE TABLE batorment_v3.analysis_queue (
    id SERIAL PRIMARY KEY,
    raid_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    skip_reason TEXT,
    platform VARCHAR(20) NOT NULL DEFAULT 'youtube',
    video_id VARCHAR(255) NOT NULL,
    CONSTRAINT analysis_queue_platform_video_id_key UNIQUE (platform, video_id),
    CONSTRAINT analysis_queue_platform_check CHECK (platform IN ('youtube', 'bilibili'))
);

-- Public correction suggestions for a video analysis (no login). The local
-- edit screen lists pending rows and marks them resolved.
CREATE TABLE batorment_v3.video_feedback (
    id SERIAL PRIMARY KEY,
    video_id VARCHAR(255) NOT NULL,
    raid_id VARCHAR(10) NOT NULL,
    user_comment TEXT NOT NULL,
    client_ip VARCHAR(64) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- What the worker read from one video against the published ranking: the
-- decision, the stated scores, the students seen and the entries that fit,
-- with the lineup frames as public image URLs. The local review screen lists
-- these; verifying the video's analysis removes its row.
CREATE TABLE batorment_v3.video_review (
    id SERIAL PRIMARY KEY,
    video_id VARCHAR(255) NOT NULL,
    raid_id VARCHAR(10) NOT NULL,
    proposal JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT video_review_video_id_raid_id_unique UNIQUE (video_id, raid_id)
);

CREATE INDEX idx_students_name ON batorment_v3.students(name_ko);
CREATE INDEX idx_students_name_ja ON batorment_v3.students(name_ja);
CREATE INDEX idx_students_name_en ON batorment_v3.students(name_en);
CREATE INDEX idx_students_name_zh ON batorment_v3.students(name_zh);
CREATE INDEX idx_students_details ON batorment_v3.students USING GIN (detail);
CREATE INDEX idx_presents_name ON batorment_v3.presents(name_ko);
CREATE INDEX idx_presents_tags ON batorment_v3.presents USING GIN (tags);
CREATE INDEX idx_i18n_category ON batorment_v3.i18n(category);
CREATE INDEX idx_youtube_analysis_raid_id ON batorment_v3.youtube_analysis(raid_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_youtube_analysis_video_raid ON batorment_v3.youtube_analysis(video_id, raid_id) WHERE deleted_at IS NULL;

ALTER TABLE batorment_v3.video_feedback ENABLE ROW LEVEL SECURITY;
