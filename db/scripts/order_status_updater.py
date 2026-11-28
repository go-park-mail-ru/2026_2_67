#!/usr/bin/python3
"""
Скрипт предназначен для обновления статуса заказа.

Запуск: каждые 1 минут.
Логика изменения:
* Обрабатывается -> В пути: если now() - updated_at >= 1 мин.
* В пути -> Ожидает в пункте выдачи: если now() - updated_at >= 1 мин.
* Ожидает в пункте выдачи -> Получен: если now() - updated_at >= 1 мин.

Да, это супер долгий скрипт который бежит по каждой строчке в таблице order, но и нагрузка не супер большая:
1 / (1 * 60) ~ 0.0167 RPS.
"""

import os
from pathlib import Path
from datetime import datetime
import psycopg


STATUS_CHANGE_QUERY = """
WITH next_status AS (
    SELECT
        o.id AS order_id,
        next_status.id AS next_status_id
    FROM "order" AS o
    JOIN order_status AS current_status
        ON current_status.id = o.order_status_id
    JOIN order_status AS next_status
        ON next_status.status = CASE current_status.status
            WHEN 'Обрабатывается' THEN 'В пути'
            WHEN 'В пути' THEN 'Ожидает в пункте выдачи'
            WHEN 'Ожидает в пункте выдачи' THEN 'Получен'
        END
    WHERE
        (current_status.status = 'Обрабатывается'
            AND o.updated_at <= NOW() - INTERVAL '1 minutes')
        OR (current_status.status = 'В пути'
            AND o.updated_at <= NOW() - INTERVAL '1 minutes')
        OR (current_status.status = 'Ожидает в пункте выдачи'
            AND o.updated_at <= NOW() - INTERVAL '1 minutes')
)
UPDATE "order" AS o
SET order_status_id = next_status.next_status_id
FROM next_status
WHERE o.id = next_status.order_id;
"""


def load_dotenv() -> None:
    """Загружает значения из корневого .env, не перезаписывая окружение."""
    env_path = next(
        (
            parent / ".env"
            for parent in Path(__file__).resolve().parents
            if (parent / ".env").is_file()
        ),
        None,
    )
    if env_path is None:
        return

    for line in env_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), value.strip().strip("'\""))


def main() -> None:
    load_dotenv()

    for key in sorted(os.environ):
        if key.startswith("DB_"):
            print(f"{key}={os.environ[key]}", flush=True)

    print("Timestamp:", datetime.now().strftime("%d.%m.%Y %H:%M:%S"))

    connection_params = {
        "user": os.environ["DB_USER"],
        "password": os.environ["DB_PASSWORD"],
        "host": os.environ["DB_HOST"],
        "port": os.environ["DB_PORT"],
        "dbname": os.environ["DB_NAME"],
    }

    with psycopg.connect(**connection_params) as connection:
        with connection.cursor() as cursor:
            cursor.execute(STATUS_CHANGE_QUERY)
            print(f"Обновлено статусов заказов: {cursor.rowcount}")


if __name__ == "__main__":
    main()
