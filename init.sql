-- Crear base de datos
CREATE DATABASE IF NOT EXISTS classifier_db;
USE classifier_db;

-- Tabla de bases registradas para escaneo
CREATE TABLE databases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    host VARCHAR(100) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Historial de ejecuciones de escaneos
CREATE TABLE scan_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    database_id INT NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (database_id) REFERENCES databases(id)
);

-- Resultados detallados por escaneo
CREATE TABLE scan_results (
    id INT AUTO_INCREMENT PRIMARY KEY,
    scan_id INT NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    column_name VARCHAR(100) NOT NULL,
    info_type VARCHAR(50) NOT NULL,
    FOREIGN KEY (scan_id) REFERENCES scan_history(id)
);

CREATE TABLE classification_rules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type_name VARCHAR(50) NOT NULL,
    regex VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initital rules insertion
INSERT INTO classification_rules (type_name, regex) VALUES
('EMAIL_ADDRESS', '(?i)email'),
('USERNAME', '(?i)user(name)?'),
('CREDIT_CARD_NUMBER', '(?i)(credit.*card|card.*number)'),
('FIRST_NAME', '(?i)first.*name'),
('LAST_NAME', '(?i)last.*name'),
('IP_ADDRESS', '(?i)ip(_address)?');
