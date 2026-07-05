CREATE DATABASE IF NOT EXISTS teacher_platform
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE teacher_platform;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS training_record;
DROP TABLE IF EXISTS training;
DROP TABLE IF EXISTS appeal;
DROP TABLE IF EXISTS teacher;

SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE teacher (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Primary key',
  user_id VARCHAR(50) NOT NULL COMMENT 'Employee number or unified identity account',
  wechat_openid VARCHAR(100) NULL COMMENT 'Wechat openid for mini program login',
  cas_account VARCHAR(100) NULL COMMENT 'CAS account for admin login',
  name VARCHAR(20) NOT NULL COMMENT 'Name',
  college VARCHAR(50) NOT NULL COMMENT 'College or unit',
  department VARCHAR(50) NULL COMMENT 'Department',
  phone VARCHAR(20) NULL COMMENT 'Masked phone number',
  email VARCHAR(100) NULL COMMENT 'Email',
  role VARCHAR(30) NOT NULL DEFAULT 'teacher' COMMENT 'teacher, party_admin, or school_admin',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_teacher_user_id (user_id),
  UNIQUE KEY uk_teacher_wechat_openid (wechat_openid),
  UNIQUE KEY uk_teacher_cas_account (cas_account),
  KEY idx_teacher_college (college),
  KEY idx_teacher_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Teacher and admin account profile';

CREATE TABLE appeal (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Treehole appeal id',
  teacher_id BIGINT NULL COMMENT 'Related teacher id; nullable for anonymous appeals',
  anonymous_type TINYINT NOT NULL DEFAULT 0 COMMENT '0 real name, 1 anonymous, 2 anonymous but contactable',
  category VARCHAR(50) NOT NULL COMMENT 'Primary category',
  sub_category VARCHAR(50) NULL COMMENT 'Secondary category',
  influence_scope TINYINT NOT NULL DEFAULT 0 COMMENT '0 personal, 1 team, 2 college, 3 school',
  emergency_level TINYINT NOT NULL DEFAULT 0 COMMENT '0 normal, 1 urgent, 2 critical',
  description TEXT NOT NULL COMMENT 'Appeal content',
  expected_method TINYINT NOT NULL DEFAULT 0 COMMENT 'Expected handling method',
  contact_way VARCHAR(100) NULL COMMENT 'Callback contact',
  attachment_url VARCHAR(255) NULL COMMENT 'Attachment URL list',
  status TINYINT NOT NULL DEFAULT 0 COMMENT '0 pending, 1 processing, 2 feedback, 3 evaluated, 4 archived',
  handler_unit VARCHAR(50) NULL COMMENT 'Handling unit',
  handler_id BIGINT NULL COMMENT 'Handler id',
  handle_content TEXT NULL COMMENT 'Handling feedback',
  satisfaction TINYINT NULL COMMENT 'Satisfaction score',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_appeal_teacher_id (teacher_id),
  KEY idx_appeal_status (status),
  KEY idx_appeal_category (category, sub_category),
  KEY idx_appeal_create_time (create_time),
  CONSTRAINT fk_appeal_teacher FOREIGN KEY (teacher_id) REFERENCES teacher (id)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Teacher treehole appeal';

CREATE TABLE training (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Training id',
  title VARCHAR(100) NOT NULL COMMENT 'Training title',
  type VARCHAR(50) NOT NULL COMMENT 'Training type',
  level TINYINT NOT NULL DEFAULT 0 COMMENT '0 school level, 1 college level',
  sponsor_unit VARCHAR(50) NULL COMMENT 'Sponsor unit',
  organizer_unit VARCHAR(50) NULL COMMENT 'Organizer unit',
  start_time DATETIME NULL COMMENT 'Start time',
  end_time DATETIME NULL COMMENT 'End time',
  location VARCHAR(100) NULL COMMENT 'Location or online URL',
  quota INT NOT NULL DEFAULT 0 COMMENT 'Quota; 0 means unlimited',
  requirements TEXT NULL COMMENT 'Enrollment requirements',
  achievement_require VARCHAR(100) NULL COMMENT 'Learning achievement requirement',
  status TINYINT NOT NULL DEFAULT 0 COMMENT '0 draft, 1 enrolling, 2 in progress, 3 ended, 4 archived',
  create_by BIGINT NULL COMMENT 'Creator id',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_training_status (status),
  KEY idx_training_time (start_time, end_time),
  KEY idx_training_type (type),
  KEY idx_training_create_by (create_by),
  CONSTRAINT fk_training_create_by FOREIGN KEY (create_by) REFERENCES teacher (id)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Ideological training activity';

CREATE TABLE training_record (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Training record id',
  training_id BIGINT NOT NULL COMMENT 'Related training id',
  teacher_id BIGINT NOT NULL COMMENT 'Related teacher id',
  apply_status TINYINT NOT NULL DEFAULT 0 COMMENT '0 pending, 1 approved, 2 rejected',
  sign_in_time DATETIME NULL COMMENT 'Sign-in time',
  study_hours DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'Learning hours',
  achievement_url VARCHAR(255) NULL COMMENT 'Learning achievement URL',
  achievement_status TINYINT NOT NULL DEFAULT 0 COMMENT '0 pending, 1 approved, 2 rejected',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_training_teacher (training_id, teacher_id),
  KEY idx_training_record_teacher_id (teacher_id),
  KEY idx_training_record_apply_status (apply_status),
  CONSTRAINT fk_training_record_training FOREIGN KEY (training_id) REFERENCES training (id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_training_record_teacher FOREIGN KEY (teacher_id) REFERENCES teacher (id)
    ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Training enrollment and learning record';

INSERT INTO teacher (id, user_id, wechat_openid, cas_account, name, college, department, phone, email, role)
VALUES
  (1, 'T20260001', 'dev-wechat-openid', NULL, 'Teacher User', 'College', 'Department', '138****0001', 'teacher@example.edu.cn', 'teacher'),
  (2, 'T20260002', NULL, 'college-admin', 'College Admin', 'College', 'Party Office', '138****0002', 'college-admin@example.edu.cn', 'party_admin'),
  (3, 'A20260001', NULL, 'school-admin', 'School Admin', 'School', 'Teacher Affairs Office', '138****0003', 'school-admin@example.edu.cn', 'school_admin');

INSERT INTO training (
  id, title, type, level, sponsor_unit, organizer_unit, start_time, end_time,
  location, quota, requirements, achievement_require, status, create_by
)
VALUES
  (1, 'Ideological Ability Training', 'workshop', 0, 'Teacher Affairs Office', 'College',
   '2026-07-08 09:00:00', '2026-07-08 11:00:00', 'Meeting Room 1', 120,
   'Open to full-time teachers', 'Submit reflection notes', 1, 3),
  (2, 'Curriculum Ideology Case Seminar', 'seminar', 1, 'College', 'Department',
   '2026-07-15 14:00:00', '2026-07-15 17:00:00', 'Online meeting', 0,
   'College teachers first', 'Submit case slides', 1, 2);

INSERT INTO training_record (
  training_id, teacher_id, apply_status, sign_in_time, study_hours,
  achievement_url, achievement_status
)
VALUES
  (1, 1, 1, '2026-07-08 08:55:00', 2.00, NULL, 0),
  (2, 1, 0, NULL, 0.00, NULL, 0);

INSERT INTO appeal (
  teacher_id, anonymous_type, category, sub_category, influence_scope,
  emergency_level, description, expected_method, contact_way, status
)
VALUES
  (1, 0, 'Teaching Support', 'Facilities', 2, 1, 'Please improve evening lighting and maintenance response.', 1, '138****0001', 0),
  (NULL, 1, 'Mental Support', 'Anonymous Feedback', 1, 0, 'Please add mental support and anonymous feedback channels for teachers.', 0, NULL, 0);
