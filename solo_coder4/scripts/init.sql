CREATE DATABASE IF NOT EXISTS solo_coder DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE solo_coder;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255),
    email VARCHAR(100),
    phone VARCHAR(20),
    avatar VARCHAR(255),
    login_type INT NOT NULL DEFAULT 1 COMMENT '1: 账号密码, 2: 微信登录',
    wechat_open_id VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    UNIQUE KEY idx_users_wechat_open_id (wechat_open_id),
    INDEX idx_username (username),
    INDEX idx_login_type (login_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE users DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users DROP INDEX IF EXISTS idx_users_phone;
