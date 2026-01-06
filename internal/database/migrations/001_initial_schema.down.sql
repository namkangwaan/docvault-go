-- Drop triggers
DROP TRIGGER IF EXISTS update_admin_users_updated_at ON admin_users;
DROP TRIGGER IF EXISTS update_files_updated_at ON files;
DROP TRIGGER IF EXISTS update_documents_updated_at ON documents;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS admin_users;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS categories;
