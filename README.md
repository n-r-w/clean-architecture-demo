# Проект на Go, построенный по принципам Clean Architecture - сервер для сохранения и получения логов
Это второй, переработанный вариант. Первая версия тут: https://github.com/n-r-w/log-server

Имеет следующие функции:
* Аутентификация
* Добавление/удаление пользователей
* Смена собственного пароля или пароля другого пользователя (только админ)
* Добавление логов
* Запрос логов по интервалу дат

Ответ на запрос логов может быть в виде:
* JSON
* JSON упакованный gzip
* JSON упакованный deflate
* Protocol Buffers упакованный gzip

Формат ответа определяется HTTP хедерами запроса

По сравнению с первой версией тут устранены лишние зависимости, особенно в presentation слое (в первой версии там полная каша). По максимуму используется изоляция модулей через интерфейсы.
В данной релизации пока нет тесткейсов (в первой версии они были) и убран доступ к сервису через web (в первой версии он есть, но сделан на скорую руку, просто чтобы посмотреть на саму возможность). Как и в первой версии тут нет DTO и сущности из домена используются на всех уровнях. 
Планируется добавить grpc интерфейс.

Описание структуры проекта:
![ScreenShot](https://github.com/n-r-w/log-server-v2/blob/main/github/info.png)

## Примеры запросов
Логин (надо сохранить полученный в ответе куки logserver для следующих запросов)

    curl --location --request POST 'localhost:8080/api/auth/login' \
    --header 'Content-Type: application/json' \    
    --data-raw '{"login": "admin", "password": "123"}'

Получить логи за период

    curl --location --request GET 'http://localhost:8080/api/private/records' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODY2MHxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXw8B2eSdqLJfQJEhsrqGnuCrf5l2_ofcwCgA0Zn0sUErg==' \
    --data-raw '{"timeFrom": "2021-04-23T14:37:36.546Z","timeTo": "2022-04-23T18:25:43.511Z"}'

Получить список пользователей

    curl --location --request GET 'http://localhost:8080/api/private/users' \
    --header 'Cookie: logserver=MTY1MTE0ODcwNHxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXwuhL1Tz50lNOOEU6N_k2oWo6wJd1ripsKVaKIJ6XxEIw=='

Состояние аутентификации

    curl --location --request GET 'http://localhost:8080/api/private/whoami' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg=='

Добавить логи

    curl --location --request POST 'http://localhost:8080/api/private/add-user' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg==' \
    --data-raw '[{"logTime": "2020-04-23T18:25:43.511Z", "level": 4, "message1": "ошибка №2"}]'

Добавить пользователя

    curl --location --request POST 'http://localhost:8080/private/add-user' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg==' \
    --data-raw '{"login": "user11","name": "user11!!!","password": "1111"}'

Сменить пароль

    curl --location --request PUT 'http://localhost:8080/api/private/change-password' \
    --header 'Content-Type: application/json' \
    --header 'Cookie: logserver=MTY1MTE0ODc0OXxEdi1CQkFFQ180SUFBUkFCRUFBQUlmLUNBQUVHYzNSeWFXNW5EQWtBQjNWelpYSmZhV1FHZFdsdWREWTBCZ0lBQVE9PXyLopILCIZS4nL8ORE6xDjmIi7aTPd77FxMBbh4apOndg==' \
    --data-raw '{"login": "user10", "password": "1111" }'

Завершить сессию

    curl --location --request DELETE 'http://localhost:8080/api/auth/close' \
    --header 'Cookie: logserver=MTY1MTE0ODkwN3xEdi1CQkFFQ180SUFBUkFCRUFBQUJQLUNBQUE9fJ6mswXN2vd3W_DpWOh7AsKYuaJiF2hd10JEUZOkKUTb'
