CREATE TABLE IF NOT EXISTS roles(
    id bigserial PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    level INT NOT NULL DEFAULT 0
);

INSERT INTO roles (name, description,level)
VALUES (
    'user', 'A user can create posts and comments', 1
);

INSERT INTO roles (name, description,level)
VALUES (
    'moderator', 'A moderator can update other users posts', 2 
);

INSERT INTO roles (name, description,level)
VALUES (
    'admin', 'A admin can update and delete other users posts', 3
);