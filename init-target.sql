-- init-target.sql
-- Target database for scanning: target_sample_db
-- Create database
CREATE DATABASE IF NOT EXISTS target_sample_db;
USE target_sample_db;

-- ----------------------------------------------------
-- TABLE: users
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
-- TABLE: addresses
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
-- TABLE: categories
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(120),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY ux_categories_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: products
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
-- TABLE: orders
-- ----------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  order_number VARCHAR(50) NOT NULL,
  total DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
  credit_card_number VARCHAR(80),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY ux_orders_number (order_number),
  INDEX idx_orders_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------------------------------
-- TABLE: order_items
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
-- TABLE: product_metadata
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
-- TABLE: activity_log
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
-- Sample data: categories (10 registros)
-- ----------------------------------------------------
INSERT INTO categories (name, slug) VALUES
('Electronics', 'electronics'),
('Books', 'books'),
('Home', 'home'),
('Clothing', 'clothing'),
('Sports', 'sports'),
('Toys', 'toys'),
('Beauty', 'beauty'),
('Groceries', 'groceries'),
('Office', 'office'),
('Garden', 'garden')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ----------------------------------------------------
-- Sample data: products (10 registros)
-- ----------------------------------------------------
INSERT INTO products (category_id, sku, name, description, price, stock)
VALUES
((SELECT id FROM categories WHERE name='Electronics'), 'SKU-EL-001', 'USB-C Charger', 'Fast charger 30W', 19.99, 150),
((SELECT id FROM categories WHERE name='Books'), 'SKU-BK-001', 'Go Programming', 'A book on Go', 39.90, 40),
((SELECT id FROM categories WHERE name='Home'), 'SKU-HM-001', 'Coffee Mug', 'Ceramic 350ml', 7.50, 200),
((SELECT id FROM categories WHERE name='Clothing'), 'SKU-CL-001', 'T-Shirt', 'Cotton T-shirt', 12.99, 300),
((SELECT id FROM categories WHERE name='Sports'), 'SKU-SP-001', 'Football', 'Size 5 official ball', 25.00, 120),
((SELECT id FROM categories WHERE name='Toys'), 'SKU-TY-001', 'Lego Set', '500 pieces building blocks', 59.99, 50),
((SELECT id FROM categories WHERE name='Beauty'), 'SKU-BT-001', 'Shampoo', '500ml bottle', 9.99, 180),
((SELECT id FROM categories WHERE name='Groceries'), 'SKU-GR-001', 'Olive Oil', '1L extra virgin', 15.50, 90),
((SELECT id FROM categories WHERE name='Office'), 'SKU-OF-001', 'Notebook', 'A4 ruled notebook', 3.25, 400),
((SELECT id FROM categories WHERE name='Garden'), 'SKU-GD-001', 'Garden Hose', '10m flexible hose', 22.75, 75)
ON DUPLICATE KEY UPDATE sku = VALUES(sku);

-- ----------------------------------------------------
-- Sample data: users (10 registros)
-- ----------------------------------------------------
INSERT INTO users (username, useremail, first_name, last_name, phone, ip_address)
VALUES
('jmendez', 'julian.mendez@example.com', 'Julian', 'Mendez', '+57-300-0000000', '192.168.1.100'),
('alice', 'alice@example.com', 'Alice', 'Smith', '+1-555-0100', '10.0.0.5'),
('bcarter', 'bcarter@example.com', 'Bob', 'Carter', '+1-555-0101', '192.168.1.101'),
('cjohnson', 'cjohnson@example.com', 'Carol', 'Johnson', '+1-555-0102', '192.168.1.102'),
('ddavis', 'ddavis@example.com', 'David', 'Davis', '+1-555-0103', '192.168.1.103'),
('emartin', 'emartin@example.com', 'Emma', 'Martin', '+1-555-0104', '192.168.1.104'),
('fflores', 'fflores@example.com', 'Felipe', 'Flores', '+57-310-0000001', '192.168.1.105'),
('ggarcia', 'ggarcia@example.com', 'Gabriela', 'Garcia', '+57-310-0000002', '192.168.1.106'),
('hlopez', 'hlopez@example.com', 'Hugo', 'Lopez', '+57-310-0000003', '192.168.1.107'),
('imoreno', 'imoreno@example.com', 'Isabel', 'Moreno', '+57-310-0000004', '192.168.1.108')
ON DUPLICATE KEY UPDATE username = VALUES(username);

-- ----------------------------------------------------
-- Sample addresses (10 registros)
-- ----------------------------------------------------
INSERT INTO addresses (user_id, address_line1, city, state, postal_code, country)
VALUES
((SELECT id FROM users WHERE username='jmendez'), 'Calle 123 #45-67', 'Bogotá', 'Bogotá D.C.', '110111', 'Colombia'),
((SELECT id FROM users WHERE username='alice'), '742 Evergreen Terrace', 'Springfield', 'IL', '62704', 'USA'),
((SELECT id FROM users WHERE username='bcarter'), 'Av. 5 #12-34', 'Medellín', 'Antioquia', '050021', 'Colombia'),
((SELECT id FROM users WHERE username='cjohnson'), '123 Main St', 'New York', 'NY', '10001', 'USA'),
((SELECT id FROM users WHERE username='ddavis'), '456 Oak St', 'Los Angeles', 'CA', '90001', 'USA'),
((SELECT id FROM users WHERE username='emartin'), '789 Pine St', 'Chicago', 'IL', '60601', 'USA'),
((SELECT id FROM users WHERE username='fflores'), 'Cra 10 #20-30', 'Cali', 'Valle', '760001', 'Colombia'),
((SELECT id FROM users WHERE username='ggarcia'), 'Cl 50 #60-70', 'Barranquilla', 'Atlántico', '080001', 'Colombia'),
((SELECT id FROM users WHERE username='hlopez'), 'Cl 25 #35-45', 'Cartagena', 'Bolívar', '130001', 'Colombia'),
((SELECT id FROM users WHERE username='imoreno'), 'Av. Siempre Viva 100', 'Lima', 'Lima', '15001', 'Peru')
ON DUPLICATE KEY UPDATE address_line1 = VALUES(address_line1);

-- ----------------------------------------------------
-- Sample orders (10 registros)
-- ----------------------------------------------------
INSERT INTO orders (user_id, order_number, total, status, credit_card_number)
VALUES
((SELECT id FROM users WHERE username='jmendez'), 'ORD-1001', 123.45, 'COMPLETED', '4111-1111-1111-1111'),
((SELECT id FROM users WHERE username='alice'), 'ORD-1002', 67.89, 'PENDING', '5500-0000-0000-0004'),
((SELECT id FROM users WHERE username='bcarter'), 'ORD-1003', 45.00, 'SHIPPED', '4111-2222-3333-4444'),
((SELECT id FROM users WHERE username='cjohnson'), 'ORD-1004', 78.20, 'COMPLETED', '5500-1111-2222-3333'),
((SELECT id FROM users WHERE username='ddavis'), 'ORD-1005', 199.99, 'PENDING', '4111-3333-4444-5555'),
((SELECT id FROM users WHERE username='emartin'), 'ORD-1006', 25.50, 'CANCELLED', '5500-2222-3333-4444'),
((SELECT id FROM users WHERE username='fflores'), 'ORD-1007', 300.00, 'COMPLETED', '4111-4444-5555-6666'),
((SELECT id FROM users WHERE username='ggarcia'), 'ORD-1008', 12.75, 'SHIPPED', '5500-3333-4444-5555'),
((SELECT id FROM users WHERE username='hlopez'), 'ORD-1009', 89.99, 'PENDING', '4111-5555-6666-7777'),
((SELECT id FROM users WHERE username='imoreno'), 'ORD-1010', 150.00, 'COMPLETED', '5500-4444-5555-6666')
ON DUPLICATE KEY UPDATE order_number = VALUES(order_number);

-- ----------------------------------------------------
-- Sample order_items (10 registros)
-- ----------------------------------------------------
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
VALUES
((SELECT id FROM orders WHERE order_number='ORD-1001'), (SELECT id FROM products WHERE sku='SKU-EL-001'), 1, 19.99),
((SELECT id FROM orders WHERE order_number='ORD-1001'), (SELECT id FROM products WHERE sku='SKU-BK-001'), 1, 39.90),
((SELECT id FROM orders WHERE order_number='ORD-1002'), (SELECT id FROM products WHERE sku='SKU-HM-001'), 2, 7.50),
((SELECT id FROM orders WHERE order_number='ORD-1003'), (SELECT id FROM products WHERE sku='SKU-CL-001'), 3, 12.99),
((SELECT id FROM orders WHERE order_number='ORD-1004'), (SELECT id FROM products WHERE sku='SKU-SP-001'), 1, 25.00),
((SELECT id FROM orders WHERE order_number='ORD-1005'), (SELECT id FROM products WHERE sku='SKU-TY-001'), 1, 59.99),
((SELECT id FROM orders WHERE order_number='ORD-1006'), (SELECT id FROM products WHERE sku='SKU-BT-001'), 2, 9.99),
((SELECT id FROM orders WHERE order_number='ORD-1007'), (SELECT id FROM products WHERE sku='SKU-GR-001'), 4, 15.50),
((SELECT id FROM orders WHERE order_number='ORD-1008'), (SELECT id FROM products WHERE sku='SKU-OF-001'), 5, 3.25),
((SELECT id FROM orders WHERE order_number='ORD-1009'), (SELECT id FROM products WHERE sku='SKU-GD-001'), 1, 22.75)
ON DUPLICATE KEY UPDATE order_id = VALUES(order_id);

-- ----------------------------------------------------
-- Sample metadata (10 registros)
-- ----------------------------------------------------
INSERT INTO product_metadata (product_id, metadata)
VALUES
((SELECT id FROM products WHERE sku='SKU-EL-001'), JSON_OBJECT('color', 'black', 'warranty_months', 12)),
((SELECT id FROM products WHERE sku='SKU-BK-001'), JSON_OBJECT('language', 'English', 'pages', 320)),
((SELECT id FROM products WHERE sku='SKU-HM-001'), JSON_OBJECT('material', 'ceramic', 'capacity_ml', 350)),
((SELECT id FROM products WHERE sku='SKU-CL-001'), JSON_OBJECT('size', 'M', 'fabric', 'cotton')),
((SELECT id FROM products WHERE sku='SKU-SP-001'), JSON_OBJECT('sport', 'football', 'weight_g', 450)),
((SELECT id FROM products WHERE sku='SKU-TY-001'), JSON_OBJECT('age', '6+', 'pieces', 500)),
((SELECT id FROM products WHERE sku='SKU-BT-001'), JSON_OBJECT('type', 'shampoo', 'volume_ml', 500)),
((SELECT id FROM products WHERE sku='SKU-GR-001'), JSON_OBJECT('origin', 'Spain', 'organic', true)),
((SELECT id FROM products WHERE sku='SKU-OF-001'), JSON_OBJECT('pages', 100, 'format', 'A4')),
((SELECT id FROM products WHERE sku='SKU-GD-001'), JSON_OBJECT('length_m', 10, 'material', 'PVC'))
ON DUPLICATE KEY UPDATE metadata = VALUES(metadata);

-- ----------------------------------------------------
-- Sample activity logs (10 registros)
-- ----------------------------------------------------
INSERT INTO activity_log (entity, entity_id, action, detail)
VALUES
('product', (SELECT id FROM products WHERE sku='SKU-EL-001'), 'CREATE', 'Product created via init script'),
('user', (SELECT id FROM users WHERE username='jmendez'), 'CREATE', 'User created via init script'),
('order', (SELECT id FROM orders WHERE order_number='ORD-1001'), 'CREATE', 'Order created'),
('order', (SELECT id FROM orders WHERE order_number='ORD-1002'), 'UPDATE', 'Order status updated'),
('product', (SELECT id FROM products WHERE sku='SKU-BK-001'), 'UPDATE', 'Price updated'),
('user', (SELECT id FROM users WHERE username='alice'), 'LOGIN', 'User login success'),
('product', (SELECT id FROM products WHERE sku='SKU-HM-001'), 'DELETE', 'Product discontinued'),
('order', (SELECT id FROM orders WHERE order_number='ORD-1005'), 'CANCEL', 'Order cancelled'),
('user', (SELECT id FROM users WHERE username='ggarcia'), 'UPDATE', 'Email updated'),
('product', (SELECT id FROM products WHERE sku='SKU-SP-001'), 'STOCK', 'Stock adjusted')
ON DUPLICATE KEY UPDATE action = VALUES(action);

-- final
COMMIT;
