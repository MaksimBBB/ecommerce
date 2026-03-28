-- Очистити таблиці
TRUNCATE TABLE order_items, orders, cart_items, products, categories, users CASCADE;

-- ============================================
-- Користувачі (пароль для всіх: Test1234!)
-- ============================================
INSERT INTO users (id, email, password_hash, first_name, last_name, role) VALUES
('11111111-1111-1111-1111-111111111111', 'admin@test.com',
 '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
 'Admin', 'User', 'admin'),
('22222222-2222-2222-2222-222222222222', 'user@test.com',
 '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
 'John', 'Doe', 'customer');

-- ============================================
-- Категорії
-- ============================================
INSERT INTO categories (id, name, description) VALUES
('c1111111-1111-1111-1111-111111111111', 'Electronics', 'Electronic devices'),
('c2222222-2222-2222-2222-222222222222', 'Books', 'Books and e-books'),
('c3333333-3333-3333-3333-333333333333', 'Clothing', 'Fashion and apparel');

-- ============================================
-- Товари (мінімум 20)
-- ============================================
INSERT INTO products (name, description, price, stock, category_id) VALUES
-- Electronics
('Laptop Pro', 'High-performance laptop', 1299.99, 50, 'c1111111-1111-1111-1111-111111111111'),
('Wireless Mouse', 'Ergonomic mouse', 29.99, 200, 'c1111111-1111-1111-1111-111111111111'),
('USB-C Hub', '7-in-1 hub', 49.99, 150, 'c1111111-1111-1111-1111-111111111111'),
('Keyboard', 'Mechanical keyboard', 89.99, 100, 'c1111111-1111-1111-1111-111111111111'),
('Webcam', '1080p webcam', 59.99, 80, 'c1111111-1111-1111-1111-111111111111'),
('Headphones', 'Wireless headphones', 149.99, 120, 'c1111111-1111-1111-1111-111111111111'),
('SSD 1TB', 'Portable SSD', 119.99, 90, 'c1111111-1111-1111-1111-111111111111'),
('Monitor 27"', '4K monitor', 399.99, 60, 'c1111111-1111-1111-1111-111111111111'),
('Go Programming', 'Learn Go', 39.99, 300, 'c2222222-2222-2222-2222-222222222222'),
('Clean Architecture', 'Design patterns', 44.99, 250, 'c2222222-2222-2222-2222-222222222222'),
('Docker Deep Dive', 'Container guide', 49.99, 200, 'c2222222-2222-2222-2222-222222222222'),
('PostgreSQL Guide', 'Database book', 54.99, 180, 'c2222222-2222-2222-2222-222222222222'),
('System Design', 'Interview prep', 42.99, 220, 'c2222222-2222-2222-2222-222222222222'),
('Sneakers', 'Sport shoes', 79.99, 150, 'c3333333-3333-3333-3333-333333333333'),
('Jacket', 'Winter jacket', 129.99, 100, 'c3333333-3333-3333-3333-333333333333'),
('Cap', 'Baseball cap', 14.99, 400, 'c3333333-3333-3333-3333-333333333333'),
('Socks Pack', '5 pairs', 9.99, 600, 'c3333333-3333-3333-3333-333333333333');

SELECT 'Seed data loaded!' AS status;
