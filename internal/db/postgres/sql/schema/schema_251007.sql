-- Create schema
CREATE SCHEMA IF NOT EXISTS batorment_v3;

-- Enum for top level
CREATE TYPE top_level AS ENUM ('I', 'T', 'L'); -- Insane, Torment, Lunatic

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
    search_keyword VARCHAR(50)[] NOT NULL,
    detail JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_students_name ON batorment_v3.students(name_ko);
CREATE INDEX IF NOT EXISTS idx_students_name_ja ON batorment_v3.students(name_ja);
CREATE INDEX IF NOT EXISTS idx_students_details ON batorment_v3.students USING GIN (detail);
