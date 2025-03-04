# ЛК

1. Добавить ручки. (делаешь максимально просто в одном файле main.go и храни данные в глобальной переменной)
    1. `POST` `/user/signup` req body{"login", "password", "name", "age", ...} resp 200, "id" (uuid). 400, 500
    2. `GET` `/user/:id` resp 200, body{"name", "age", ...}. 400, 500
    3. `POST` `/user/login` req body{"login", "password", ...} resp 200, "id" (uuid). 400, 500, 403
    4. `PUT` `/user/:id` req body{"name", "age", ...} resp 200, "id" (uuid). 400, 500
    5. `DELETE` `/user/:id` resp 200, 400, 404, 500

2. слайс vs массив, структура слайса, что происходит при append
3. map, бакеты, что происходит при коллизиях, миграции