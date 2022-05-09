# Вторая версия проекта на Go, построенного по принципам Clean Architecture

Первая версия: https://github.com/n-r-w/log-server

По сравнению с первой версией тут устранены лишние зависимости, особенно в presentation слое (в первой версии там полная каша). По максимуму используется изоляция модулей через интерфейсы. 
Некоторым отступлением от классики можно считать объявление интерфейсов по месту реализации, а не использования. Но в данном случае это выглядит более логично. 
К примеру, мы заходим в каталог с юскейсами и сразу видим там их интефейс https://github.com/n-r-w/log-server-v2/blob/main/internal/usecase/usecase/interface.go 
Поскольку в Go реализация интерфейса не приводит к ссылке на пакет, где он описан, то циклических зависимостей не образуется.

В данной релизации пока нет тесткейсов (в первой версии они были) и убран доступ к сервису через web (в первой версии он есть, но сделан на скорую руку, 
просто чтобы посмотреть на саму возможность).

Планируется добавить grpc интерфейс.
