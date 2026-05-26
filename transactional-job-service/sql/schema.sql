CREATE DATABASE IF NOT EXISTS transactional_job_service
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

CREATE TABLE transactional_jobs (
    id VARCHAR(128) NOT NULL,
    callback_url VARCHAR(2048) NOT NULL,
    payload JSON NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    available_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processing_at TIMESTAMP DEFAULT NULL,
    claim_timeout INT NOT NULL DEFAULT 60,
    retry_seconds INT NOT NULL DEFAULT 10,
    retry_backoff DECIMAL(10, 2) NOT NULL DEFAULT 1.5,
    max_attempts INT NOT NULL DEFAULT 10,
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4;