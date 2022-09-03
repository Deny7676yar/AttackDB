\c gopher_corp
SET ROLE gopher;

DROP TABLE IF EXISTS departments CASCADE;
CREATE TABLE departments (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    parent_id INT NOT NULL,
    name VARCHAR(200) UNIQUE NOT NULL
);

DROP TABLE IF EXISTS positions CASCADE;
CREATE TABLE positions(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title VARCHAR(200) UNIQUE NOT NULL
);

DROP TABLE IF EXISTS employees CASCADE;
CREATE TABLE employees(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    first_name VARCHAR(200) NOT NULL,
    last_name VARCHAR(200) NOT NULL,
    phone TEXT,
    email TEXT,
    salary MONEY NOT NULL,
    manager_id INT NOT NULL,
    department INT NOT NULL REFERENCES departments(id),
    position INT NOT NULL REFERENCES positions(id),
    entry_at  DATE NOT NULL DEFAULT CURRENT_DATE,
    FOREIGN KEY (manager_id) REFERENCES employees (id)
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT employees_salary_positive_check CHECK (salary::numeric > 0)
);

CREATE INDEX
ON employees(email text_pattern_ops);