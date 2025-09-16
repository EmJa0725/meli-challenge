-- init-target.sql
-- Target database for scanning: target_sample_db
-- Create database
CREATE DATABASE IF NOT EXISTS target_sample_db;
USE target_sample_db;

-- ----------------------------------------------------
-- TABLE: users (contains personal - sensitive data)
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(100) NOT NULL,
  useremail VARCHAR(150) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  phone VARCHAR(50),
  ip_address VARCHAR(45),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY ux_users_username (username),
  INDEX idx_users_email (useremail)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: addresses (not sensitive in this context)
-- Related to users (1:N)
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS addresses (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  address_line1 VARCHAR(255),
  address_line2 VARCHAR(255),
  city VARCHAR(100),
  state VARCHAR(100),
  postal_code VARCHAR(20),
  country VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_addresses_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: categories (not sensitive)
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(120),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY ux_categories_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: products (not sensitive)
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
  id INT AUTO_INCREMENT PRIMARY KEY,
  category_id INT,
  sku VARCHAR(80) NOT NULL,
  name VARCHAR(200) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  stock INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
  UNIQUE KEY ux_products_sku (sku),
  INDEX idx_products_category (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: orders (contains possible card reference - sensitive)
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  order_number VARCHAR(50) NOT NULL,
  total DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
  credit_card_number VARCHAR(80), -- campo sensible
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY ux_orders_number (order_number),
  INDEX idx_orders_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: order_items (order details)
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS order_items (
  id INT AUTO_INCREMENT PRIMARY KEY,
  order_id INT NOT NULL,
  product_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  unit_price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
  INDEX idx_order_items_order (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: product_metadata (not sensitive)
-- JSON/BLOB for ad-hoc data
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS product_metadata (
  id INT AUTO_INCREMENT PRIMARY KEY,
  product_id INT NOT NULL,
  metadata JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
  INDEX idx_pm_product (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: activity_log (not sensitive - for auditing)
-- stores system events, not personal data
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS activity_log (
  id INT AUTO_INCREMENT PRIMARY KEY,
  entity VARCHAR(100),
  entity_id INT,
  action VARCHAR(100),
  detail TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_activity_entity (entity, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- ----------------------------------------------------
-- Sample data: categories and products
-- ----------------------------------------------------
INSERT INTO categories (name, slug) VALUES
('Electronics', 'electronics'),
('Books', 'books'),
('Home', 'home')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO products (category_id, sku, name, description, price, stock)
VALUES
((SELECT id FROM categories WHERE name='Electronics'), 'SKU-EL-001', 'USB-C Charger', 'Fast charger 30W', 19.99, 150),
((SELECT id FROM categories WHERE name='Books'), 'SKU-BK-001', 'Go Programming', 'A book on Go', 39.90, 40),
((SELECT id FROM categories WHERE name='Home'), 'SKU-HM-001', 'Coffee Mug', 'Ceramic 350ml', 7.50, 200)
ON DUPLICATE KEY UPDATE sku = VALUES(sku);


-- ----------------------------------------------------
-- Sample data: users (includes sensitive fields)
-- ----------------------------------------------------
INSERT INTO users (username, useremail, first_name, last_name, phone, ip_address)
VALUES
('jmendez', 'julian.mendez@example.com', 'Julian', 'Mendez', '+57-300-0000000', '192.168.1.100'),
('alice', 'alice@example.com', 'Alice', 'Smith', '+1-555-0100', '10.0.0.5')
ON DUPLICATE KEY UPDATE username = VALUES(username);

-- Sample addresses
INSERT INTO addresses (user_id, address_line1, city, state, postal_code, country)
VALUES
((SELECT id FROM users WHERE username='jmendez'), 'Calle 123 #45-67', 'Bogotá', 'Bogotá D.C.', '110111', 'Colombia'),
((SELECT id FROM users WHERE username='alice'), '742 Evergreen Terrace', 'Springfield', 'IL', '62704', 'USA')
ON DUPLICATE KEY UPDATE address_line1 = VALUES(address_line1);

-- ----------------------------------------------------
-- Sample data: orders and items (includes fictitious card)
-- ----------------------------------------------------
INSERT INTO orders (user_id, order_number, total, status, credit_card_number)
VALUES
((SELECT id FROM users WHERE username='jmendez'), 'ORD-1001', 123.45, 'COMPLETED', '4111-1111-1111-1111'),
((SELECT id FROM users WHERE username='alice'), 'ORD-1002', 67.89, 'PENDING', '5500-0000-0000-0004')
ON DUPLICATE KEY UPDATE order_number = VALUES(order_number);

INSERT INTO order_items (order_id, product_id, quantity, unit_price)
VALUES
((SELECT id FROM orders WHERE order_number='ORD-1001'), (SELECT id FROM products WHERE sku='SKU-EL-001'), 1, 19.99),
((SELECT id FROM orders WHERE order_number='ORD-1001'), (SELECT id FROM products WHERE sku='SKU-BK-001'), 1, 39.90),
((SELECT id FROM orders WHERE order_number='ORD-1002'), (SELECT id FROM products WHERE sku='SKU-HM-001'), 2, 7.50)
ON DUPLICATE KEY UPDATE order_id = VALUES(order_id);

-- Sample metadata
INSERT INTO product_metadata (product_id, metadata)
VALUES
((SELECT id FROM products WHERE sku='SKU-EL-001'), JSON_OBJECT('color', 'black', 'warranty_months', 12)),
((SELECT id FROM products WHERE sku='SKU-BK-001'), JSON_OBJECT('language', 'English', 'pages', 320))
ON DUPLICATE KEY UPDATE metadata = VALUES(metadata);

-- Activity logs
INSERT INTO activity_log (entity, entity_id, action, detail)
VALUES
('product', (SELECT id FROM products WHERE sku='SKU-EL-001'), 'CREATE', 'Product created via init script'),
('user', (SELECT id FROM users WHERE username='jmendez'), 'CREATE', 'User created via init script')
ON DUPLICATE KEY UPDATE action = VALUES(action);

-- final
COMMIT;
