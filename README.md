# Load generation

./app -mode=4 -maxConns=1 -minConns=1 -attackMS=200 -goroutines=15

# FOR UPDATE

В первом терминале выполним:

```sql
SELECT *
FROM employees
WHERE first_name = 'Alice'
FOR UPDATE;
```

Во втором терминале выполним:

```sql
SELECT *
FROM employees
FOR UPDATE;
```

Что наблюдаем? Почему?

# FOR SHARE

В первом терминале выполним:

```sql
SELECT *
FROM employees
WHERE first_name = 'Alice'
FOR SHARE;
```

Во втором терминале выполним:

```sql
SELECT *
FROM employees
FOR SHARE;
```

Что наблюдаем? Почему?

Во втором терминале выполним:

```sql
SELECT *
FROM employees
FOR UPDATE;
```

Что наблюдаем? Почему?

# DB Prepared Statement

Создадим Prepared Statement:

```sql
PREPARE prepEmailByName (VARCHAR) AS
    SELECT email
    FROM employees
    WHERE first_name = $1 AND last_name = $2;
```

Посмотрим, что создалось:

```sql
SELECT *
FROM pg_prepared_statements;
```

Можно также посмотреть Execution Plan:

```sql
EXPLAIN EXECUTE prepEmailByName('Alice', 'Liddell'); 
```

Выполним:

```sql
EXECUTE prepEmailByName('Alice', 'Liddell');
```

Удалим созданное Prepared Statement:

```sql
DEALLOCATE prepEmailByName;
```

Проверим, что случится с Prepared Statement при разрыве соединения и при создании нового соединения.

# PgBouncer

```bash
docker run --rm \
    -e DATABASE_URL="postgres://gopher:P@ssw0rd@172.0.0.1/database" \
    -p 5432:5432 \
    edoburu/pgbouncer
```