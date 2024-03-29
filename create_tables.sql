CREATE TABLE IF NOT EXISTS uploads (
  id int not null auto_increment,
  vault_name varchar(255),
  filename varchar(255),
  created_at varchar(255),
  deleted_at varchar(255),
  location varchar(255),
  session_id varchar(255),
  checksum varchar(255),
  archive_id varchar(255),
  status int,
  PRIMARY KEY(id),
  INDEX filename_idx (filename),
  INDEX created_at_idx (created_at)
) ENGINE=InnoDB CHARACTER SET=utf8mb4;

CREATE TABLE IF NOT EXISTS uploaded_segments (
  id int not null auto_increment,
  upload_id int,
  segment_num int,
  checksum varchar(255),
  created_at varchar(255),
  PRIMARY KEY(id),
  FOREIGN KEY (upload_id) REFERENCES uploads(id) ON DELETE CASCADE
) ENGINE=InnoDB CHARACTER SET=utf8mb4;

CREATE TABLE IF NOT EXISTS downloads (
  id int not null auto_increment,
  vault_name varchar(255),
  job_id varchar(255),
  archive_id varchar(255),
  sns_topic varchar(255),
  created_at varchar(255),
  updated_at varchar(255),  
  status int,
  PRIMARY KEY(id),
  INDEX created_at_idx (created_at),
  INDEX archive_id_idx (archive_id)
) ENGINE=InnoDB CHARACTER SET=utf8mb4;

CREATE TABLE IF NOT EXISTS downloaded_segments (
  id int not null auto_increment,
  download_id int,
  bytes_range varchar(255),  
  created_at varchar(255), 
  PRIMARY KEY(id),
  FOREIGN KEY (download_id) REFERENCES downloads(id) ON DELETE CASCADE
) ENGINE=InnoDB CHARACTER SET=utf8mb4;
