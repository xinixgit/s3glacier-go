CREATE TABLE IF NOT EXISTS uploads (
  id int not null auto_increment,
  vault_name varchar(255),
  filename varchar(255),
  created_at varchar(255),
  location varchar(255),
  session_id varchar(255),
  checksum varchar(255),
  archive_id varchar(255),
  status int,
  PRIMARY KEY(id)
) ENGINE=InnoDB CHARACTER SET=utf8mb4;

CREATE TABLE IF NOT EXISTS upload_segments (
  id int not null auto_increment,
  upload_id int,
  segment_num int,
  checksum varchar(255),
  created_at varchar(255),
  PRIMARY KEY(id),
  FOREIGN KEY (upload_id) REFERENCES uploads(id) ON DELETE CASCADE
) ENGINE=InnoDB CHARACTER SET=utf8mb4;