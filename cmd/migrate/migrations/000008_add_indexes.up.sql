CREATE INDEX idx_comments_content ON comments USING gin (content gin_trgm_ops);

CREATE INDEX idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX idx_posts_tags ON posts USING gin (tags);

CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_posts_user_id ON posts (user_id);
CREATE INDEX idx_comments_post_id ON comments (post_id);

