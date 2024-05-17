CREATE SCHEMA IF NOT EXISTS s3g;

CREATE TABLE IF NOT EXISTS s3g.uploads (
  id SERIAL PRIMARY KEY,
  vault_name varchar(255),
  filename varchar(255),
  created_at varchar(255),
  deleted_at varchar(255),
  location varchar(255),
  session_id varchar(255),
  checksum varchar(255),
  archive_id varchar(255),
  status int
);

CREATE INDEX IF NOT EXISTS filename_idx ON s3g.uploads(filename);
CREATE INDEX IF NOT EXISTS created_at_idx ON s3g.uploads(created_at);

CREATE TABLE IF NOT EXISTS s3g.uploaded_segments (
  id SERIAL PRIMARY KEY,
  upload_id int REFERENCES s3g.uploads (id) ON DELETE CASCADE,
  segment_num int,
  checksum varchar(255),
  created_at varchar(255)
);

CREATE TABLE IF NOT EXISTS s3g.downloads (
  id SERIAL PRIMARY KEY,
  vault_name varchar(255),
  job_id varchar(255),
  archive_id varchar(255),
  sns_topic varchar(255),
  created_at varchar(255),
  updated_at varchar(255),
  status int
);

CREATE INDEX IF NOT EXISTS dl_created_at_idx ON s3g.downloads(created_at);
CREATE INDEX IF NOT EXISTS dl_archive_id_idx ON s3g.downloads(archive_id);

CREATE TABLE IF NOT EXISTS s3g.downloaded_segments (
  id SERIAL PRIMARY KEY,
  download_id int REFERENCES s3g.downloads(id) ON DELETE CASCADE,
  bytes_range varchar(255),
  created_at varchar(255)
);
