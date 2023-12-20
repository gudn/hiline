# hiline

Программа для создания временных линий. Хранит все данные локально в обычной
папке в простом виде, что отлично подходит для синхронизации через Git.

## Модель

Система оперирует документами, которые занимают свое место или отрезок на
временной линии. Временных линий может быть несколько, это называется группа.
Каждый документ обязан быть в какой-то группе, а сами группы могут быть
вложенными.

## Управление

Для создания временных линий используется кнопка `Add group` или через `Ctrl-I`.
Нас попросят ввести имя для новой линии, вложенность достигается через `/`.
Чтобы удалить линию надо кликнуть правой кнопкой на название слева.

Перемещаться во времени можно мышкой, маштаб через колесико с зажатым `Shift`.

Добавить новый документ можно двойным кликом в нужном месте. Выделив документ
(будет другого цвета), его можно перемещать (в том числе между линиями) или
удалить. Если создать новую линию, когда есть выделенный документ, создатся
вложенная линия с таким же именем.

По ПКМ на документе открывается вкладка с редактированием. Там можно поменять
название или содержимое документа.

## Данные

Все данные автоматически сохраняются на диск, по одному файлу на каждый
документ. Группы нигде не хранятся, существуют только те группы, в которых
есть хоть один документ.
