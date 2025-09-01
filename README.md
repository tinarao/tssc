# tssc

<p align="center">
    <img src="https://skillicons.dev/icons?i=go" />
</p>

CLI-приложение для управления shadowsocks прокси-соединениями.

## Оригинальный пакет

`tssc` использует [outline-sdk](https://github.com/Jigsaw-Code/outline-sdk) от Jigsaw-Code, который предоставляет базовый функционал для работы с Shadowsocks.

Утилита требует root-права для работы с сетевыми соединениями. 

## Команды

- `connect <alias>` - установить проки-соединение используя конфиг по алиасу
- `list` - вывести список сохранённых конфигов в формате `<alias> :: <ss://>`
- `add <alias> <ss://url>` - сохранить конфиг под алиасом

## TODO

- [ ] Install script
