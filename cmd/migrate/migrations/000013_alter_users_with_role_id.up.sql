-- ALTER TABLE users
-- ADD COLUMN role_id BIGINT REFERENCES roles(id);

-- 問題：既有資料的 role_id 會是 NULL:

-- 加欄位後舊資料全部是 NULL，無法直接加 NOT NULL
-- 解法：先允許 NULL，UPDATE 補值，再加 NOT NULL

-- 1. 先允許 NULL 加欄位
ALTER TABLE users
ADD COLUMN role_id BIGINT REFERENCES roles(id);

-- 2. 補上所有既有資料的值
UPDATE users SET role_id = (
    SELECT id FROM roles 
    WHERE name = 'user'
);

-- 3. 加上 NOT NULL
ALTER TABLE users
ALTER COLUMN role_id SET NOT NULL;