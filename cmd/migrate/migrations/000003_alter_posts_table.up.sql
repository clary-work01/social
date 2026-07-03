ALTER TABLE posts 
ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id);