-- init.sql
-- Create database
CREATE DATABASE IF NOT EXISTS classifier_db;
USE classifier_db;

-- Table of registered databases for scanning
CREATE TABLE `external_databases` (
    id INT AUTO_INCREMENT PRIMARY KEY,
    host VARCHAR(100) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Scan executions history
CREATE TABLE scan_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    database_id INT NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    FOREIGN KEY (database_id) REFERENCES `external_databases`(id)
);

-- Detailed results per scan (now includes schema_name)
CREATE TABLE scan_results (
    id INT AUTO_INCREMENT PRIMARY KEY,
    scan_id INT NOT NULL,
    schema_name VARCHAR(100) NOT NULL,
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

-- Initial rules insertion
INSERT INTO classification_rules (type_name, regex) VALUES
-- Personal Information (PII)
('FIRST_NAME', '(?i)^first(_?name)?$'),
('LAST_NAME', '(?i)^last(_?name)?$'),
('DATE_OF_BIRTH', '(?i)^(dob|date[_ ]?of[_ ]?birth|birth[_ ]?date)$'),
('GENDER', '(?i)^(gender|sex)$'),
('SSN', '(?i)^(ssn|social[_ ]?security[_ ]?number|national[_ ]?id)$'),

-- Contact Information
('EMAIL_ADDRESS', '(?i)email'),
('PHONE_NUMBER', '(?i)^(phone|mobile|contact[_ ]?number)$'),
('ADDRESS', '(?i)^address(_.*)?$'),
('POSTAL_CODE', '(?i)^(postal|zip)_?code$'),

-- Authentication / Security Data
('USERNAME', '(?i)^user(name)?$'),
('PASSWORD', '(?i)^password$'),
('SECURITY_QUESTION', '(?i)^security[_ ]?(question|answer)$'),
('API_KEY', '(?i)^(api[_ ]?key|auth[_ ]?token|secret)$'),

-- Financial Information (PCI)
('CREDIT_CARD_NUMBER', '(?i)^(credit[_ ]?card(_?number)?|card[_ ]?number)$'),
('BANK_ACCOUNT', '(?i)^(account[_ ]?number|iban|acct)$'),
('ROUTING_NUMBER', '(?i)^routing[_ ]?number$'),
('SWIFT_CODE', '(?i)^swift[_ ]?code$'),

-- Technical Identifiers
('IP_ADDRESS', '(?i)^ip(_address)?$'),
('MAC_ADDRESS', '(?i)^mac[_ ]?address$'),
('HOSTNAME', '(?i)^host[_ ]?name$');