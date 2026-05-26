-- Create schema
CREATE SCHEMA IF NOT EXISTS batorment_v3;

-- Enum for top level
CREATE TYPE top_level AS ENUM ('I', 'T', 'L'); -- Insane, Torment, Lunatic

-- Enum for analysis type
CREATE TYPE analysis_type AS ENUM ('ai', 'user');

create table batorment_v3.youtube_analysis (
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

-- Table for contents in Blue Archive
create table batorment_v3.contents (
  content_id VARCHAR(10) NOT NULL,
  title VARCHAR(200) NOT NULL,
  top_level top_level NOT NULL,
  search_keyword VARCHAR(200) NOT NULL,
  start_date TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

-- Table for students in Blue Archive
CREATE TABLE batorment_v3.students (
    student_id INTEGER NOT NULL PRIMARY KEY,
    name_ko VARCHAR(50) NOT NULL,
    name_ja VARCHAR(50) NOT NULL,
    name_en VARCHAR(50) NOT NULL DEFAULT '',
    name_zh VARCHAR(50) NOT NULL DEFAULT '',
    search_keyword VARCHAR(50)[] NOT NULL,
    detail JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Table for presents in Blue Archive
CREATE TABLE batorment_v3.presents (
    present_id INTEGER NOT NULL PRIMARY KEY,
    name_ko VARCHAR(50) NOT NULL,
    rarity VARCHAR(10) NOT NULL,
    tags VARCHAR(50)[] NOT NULL,
    exp_value INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Table for i18n translations (school, club, etc.)
CREATE TABLE batorment_v3.i18n (
    category VARCHAR(20) NOT NULL,
    key VARCHAR(50) NOT NULL,
    name_ko VARCHAR(100) NOT NULL,
    name_ja VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    name_zh VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (category, key)
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_students_name ON batorment_v3.students(name_ko);
CREATE INDEX IF NOT EXISTS idx_students_name_ja ON batorment_v3.students(name_ja);
CREATE INDEX IF NOT EXISTS idx_students_name_en ON batorment_v3.students(name_en);
CREATE INDEX IF NOT EXISTS idx_students_name_zh ON batorment_v3.students(name_zh);
CREATE INDEX IF NOT EXISTS idx_students_details ON batorment_v3.students USING GIN (detail);
CREATE INDEX IF NOT EXISTS idx_presents_name ON batorment_v3.presents(name_ko);
CREATE INDEX IF NOT EXISTS idx_presents_tags ON batorment_v3.presents USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_i18n_category ON batorment_v3.i18n(category);

