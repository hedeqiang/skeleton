-- 区块链交换项目数据库初始化脚本

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建 hello 表 (示例表)
CREATE TABLE IF NOT EXISTS hello_records (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_hello_records_created_at ON hello_records(created_at);

-- 插入初始数据
INSERT INTO users (username, email, password_hash) VALUES 
    ('admin', 'admin@example.com', '$2a$10$example.hash.value'),
    ('user1', 'user1@example.com', '$2a$10$example.hash.value')
ON CONFLICT (username) DO NOTHING;

INSERT INTO hello_records (message) VALUES 
    ('Hello, World!'),
    ('Welcome to blockchain swap!')
ON CONFLICT DO NOTHING; 