CREATE DATABASE IF NOT EXISTS personal_blog
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE personal_blog;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  name VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(16) NOT NULL DEFAULT 'user',
  PRIMARY KEY (id),
  UNIQUE KEY uni_users_email (email),
  KEY idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(80) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uni_categories_name (name),
  UNIQUE KEY uni_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS articles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  title VARCHAR(180) NOT NULL,
  slug VARCHAR(220) NOT NULL,
  summary TEXT NULL,
  cover_url VARCHAR(500) NULL,
  content LONGTEXT NOT NULL,
  tags TEXT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'draft',
  published_at DATETIME(3) NULL,
  category_id BIGINT UNSIGNED NULL,
  author_id BIGINT UNSIGNED NOT NULL,
  likes_count BIGINT NOT NULL DEFAULT 0,
  favorites_count BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uni_articles_slug (slug),
  KEY idx_articles_status (status),
  KEY idx_articles_category_id (category_id),
  KEY idx_articles_author_id (author_id),
  KEY idx_articles_published_at (published_at),
  KEY idx_articles_search (title, summary(191), tags(191)),
  CONSTRAINT fk_articles_category
    FOREIGN KEY (category_id) REFERENCES categories(id)
    ON UPDATE CASCADE ON DELETE SET NULL,
  CONSTRAINT fk_articles_author
    FOREIGN KEY (author_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  article_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  body TEXT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'published',
  PRIMARY KEY (id),
  KEY idx_comments_article_id (article_id),
  KEY idx_comments_user_id (user_id),
  KEY idx_comments_status (status),
  CONSTRAINT fk_comments_article
    FOREIGN KEY (article_id) REFERENCES articles(id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_comments_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  article_id BIGINT UNSIGNED NOT NULL,
  type VARCHAR(16) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY ux_user_article_type (user_id, article_id, type),
  KEY idx_reactions_article_id (article_id),
  CONSTRAINT fk_reactions_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_reactions_article
    FOREIGN KEY (article_id) REFERENCES articles(id)
    ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO categories (name, slug, created_at, updated_at) VALUES
  ('Go', 'go', NOW(3), NOW(3)),
  ('Vue', 'vue', NOW(3), NOW(3)),
  ('数据库', 'database', NOW(3), NOW(3)),
  ('前端', 'frontend', NOW(3), NOW(3));

