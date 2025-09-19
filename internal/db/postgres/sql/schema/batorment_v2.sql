-- Create schema
CREATE SCHEMA IF NOT EXISTS batorment_v2;

-- Named users table
CREATE TABLE IF NOT EXISTS batorment_v2.videos (
    id SERIAL PRIMARY KEY,
    raid_id VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    youtube_url VARCHAR(200) NOT NULL,
    score INTEGER NOT NULL,
    ai_generated_detail JSONB,
    parsed_detail JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE (youtube_url)
);

-- Raids table  
CREATE TABLE IF NOT EXISTS batorment_v2.raids (
    raid_id VARCHAR(20) NOT NULL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    short_name VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    top_level VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Students table
CREATE TABLE IF NOT EXISTS batorment_v2.students (
    student_id INTEGER NOT NULL PRIMARY KEY,
    name_ko VARCHAR(50) NOT NULL,
    name_ja VARCHAR(50) NOT NULL,
    search_keyword VARCHAR(50)[] NOT NULL,
    detail JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_students_name ON batorment_v2.students(name_ko);
CREATE INDEX IF NOT EXISTS idx_students_name_ja ON batorment_v2.students(name_ja);
CREATE INDEX IF NOT EXISTS idx_students_details ON batorment_v2.students USING GIN (detail);
CREATE INDEX IF NOT EXISTS idx_videos_raid_id ON batorment_v2.videos(raid_id);
