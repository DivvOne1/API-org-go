-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INT NULL REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_departments_root_name
    ON departments (name)
    WHERE parent_id IS NULL;

CREATE UNIQUE INDEX ux_departments_parent_name
    ON departments (parent_id, name)
    WHERE parent_id IS NOT NULL;

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    department_id INT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_employees_department_id ON employees (department_id);
CREATE INDEX idx_departments_parent_id ON departments (parent_id);

-- +goose Down
DROP TABLE employees;
DROP INDEX IF EXISTS ux_departments_parent_name;
DROP INDEX IF EXISTS ux_departments_root_name;
DROP INDEX IF EXISTS idx_departments_parent_id;
DROP TABLE departments;
