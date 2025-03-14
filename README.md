# NetMafia 3.0

**NetMafia 3.0** — это Docker-решение для развертывания серверной части игры в мафию. Проект включает в себя два основных сервиса: базу данных PostgreSQL и основное приложение (mafia_app), которое работает в контейнере.

## Особенности

- **Dockerized**: Полностью контейнеризованное приложение для лёгкого развертывания и масштабирования.
- **PostgreSQL**: Использование PostgreSQL для хранения данных.

## Требования

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Установка и запуск

1. **Клонируйте репозиторий**

   Если вы работаете через GitHub Web UI, можно воспользоваться встроенным редактором или создать pull request для внесения изменений. Если же вы работаете локально, выполните:

   ```bash
   git clone https://github.com/Ggosh1/NetMafia3.0.git
   cd NetMafia3.0
```

**Настройте переменные окружения**

Создайте файл `.env` в корневой директории проекта и укажите в нём необходимые переменные (этот файл не должен попадать в публичный репозиторий):

bash

Копировать
```
#.env
POSTGRES_USER=your_postgres_user
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=your_database_name

```

> Если вы используете GitHub Actions, секреты можно передавать напрямую через раздел `env` в workflow-файле.

**Запустите контейнеры**

Для сборки и запуска приложения выполните:
```bash
`docker compose up -d --build`
```


Эта команда соберёт образы и запустит сервисы в фоне. Приложение запускается на 8080 порту по умолчанию.



## Остановка приложения

Чтобы остановить и удалить контейнеры, выполните:


```bash
docker compose down
```



# Описание ролей

## Доктор (Doctor)
- **Описание**: Ты можешь выбрать игрока для защиты. Выбранный игрок не будет убит этой ночью.  
- **Команда**: Жители  
- **Аура**: Хороший  

---

## Провидец (Seer)
- **Описание**: Каждую ночь ты можешь выбрать игрока, чтобы узнать его роль.  
- **Команда**: Жители  

---

## Крикун (Loudmouth)
- **Описание**: Ты можешь выбрать игрока, роль которого будет раскрыта всем после твоей смерти.  
- **Команда**: Жители  

---

## Альфа оборотень (Alpha werewolf)
- **Описание**: Твой ночной голос считается за два. В течение дня ты можешь отправлять личные сообщения другим оборотням, которые могут видеть только они.  
- **Команда**: Оборотни  

---

## Обычный оборотень (Werewolf)
- **Описание**: Каждую ночь ты можешь голосовать с оборотнями за игрока, которого необходимо убить.  
- **Команда**: Оборотни  

---

## Житель (Villager)
- **Описание**: Ты обычный житель без каких-либо способностей.  
- **Команда**: Жители  

---

## Шут (Fool)
- **Описание**: Твоя цель — быть казнённым жителями. Если тебя казнят — ты победишь.  
- **Команда**: Одиночка  
