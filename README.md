# mssql-ci
Utility for copying procedures and view from MSSQL to files

**Использование:**

`$ mssql-ci -cmd=pull [-env="custom env file"]`

Параметр *cmd* (обязательный) - выполняемая команда

Значения:

* *pull* - забирает данные с MSSQL сервера 

Параметр *env* (необязательный) - имя файла с параметрами подключения к MSSQL

Значение по умолчанию .env

**Описание env файла:**

* SQLSERVER - имя MSSQL сервера
* PORT - порт MSSQL сервера
* DATABASE - база данных
* USERNAME - имя пользователя для подключения
* PASSWORD - пароль
* FILESTORE - путь к каталогу для хранения файлов процедур

